package potential

import (
	"fmt"
	"sync"
)

// All methods on a network that relate to growing are here.

/*
Grow is a general growth that encompasses all growth methods.

It adds neurons, adds new synapses, prunes old neurons, and strengthens synapses that have
fired a lot.
*/
func (network *Network) Grow(neuronsToAdd, defaultNeuronSynapses, synapsesToAdd int) {
	network.GrowRandomNeurons(neuronsToAdd, defaultNeuronSynapses)
	network.GrowRandomSynapses(synapsesToAdd)
}

/*
GrowPathBetween will make a path between two neurons.

The desired goal is for there to be a path, or series of paths, from startCell to endCell.

1. We start in at the startCell and traverse its axons.
2. At each layer, we check to see if any of the start axons are connected to the
end dendrites.
3. We count the connected synapses as we go.
4. Upon reaching minSynapses, or maxHops, or nowhere, we stop.
	- if reached minSynapses, just really stop and return 0 synapses added
	- if reached maxHops, continue.
5. We need to add enough synapses to reach minSynapses, such that these are the minimum
number of synapses that connect from startCell's tree to endCell.

After `maxHops`, if there are not minSynapses, we create synapses at that layer.
*/
func (network *Network) GrowPathBetween(startCell, endCell CellID, minSynapses int) (synapsesToEnd map[SynapseID]bool, synapsesAdded map[SynapseID]bool) {
	// these are the synapses we found that are on the path from the startCell,
	// and attach directly to an endCell at the dendrite
	synapsesToEnd = make(map[SynapseID]bool)
	// any new synapses we create if there are not enough in the network that attach
	// to the end cell
	synapsesAdded = make(map[SynapseID]bool)

	maxHops := network.maxHops

	mux := sync.Mutex{}
	hops := 0
	var wg sync.WaitGroup
	ch := make(chan SynapseID)
	alreadyWalked := make(map[CellID]bool)

	// walk traverses the axons and see if any synapse leads to the end cell.
	// hops is the layer we are on, *copied* on pass.
	// this is a fan-out kind of traversal through a tree of cells and synapses.
	// Note: need to fully declare it before assigning it, apparently because the
	// runtime needs this to compile a recursive function.
	// TODO: refactor to be more elegant and use no mutexes or wait groups
	// once i better understand goroutines and channels
	var walk func(cellID CellID)
	walk = func(cellID CellID) {
		mux.Lock()
		hops++
		totalSynapsesFound := len(synapsesToEnd)
		if _, already := alreadyWalked[cellID]; already {
			mux.Unlock()
			return
		}
		alreadyWalked[cellID] = true
		mux.Unlock()

		if hops >= maxHops || totalSynapsesFound >= minSynapses {
			return
		}
		wg.Add(1)
		// look at the next cells in the axon chain from this one, to see
		// if any are the endCell then send it down the channel.
		go func() {
			for axonSynapseID := range network.Cells[cellID].AxonSynapses {
				s, exists := network.Synapses[axonSynapseID]
				if !exists {
					fmt.Println("warn: cannot grow path because synapse axon does not exist",
						axonSynapseID, "from cell=", cellID, network.Cells[cellID].Tag)
					continue
				}
				receiverCellID := s.ToNeuronDendrite

				if receiverCellID == endCell {
					ch <- axonSynapseID
				}
				// also walk the axons of this cell if the synapse is
				// excitatory.
				if s.Millivolts > 0 {
					walk(receiverCellID)
				}
			}
			wg.Done()
		}()
	}

	// receive the connected synapses.
	// channel is closed upstream when we reach the end (? how ?)
	go func() {
		walk(startCell)
		wg.Wait()
		close(ch)
	}()

	for synapseToOutputCell := range ch {
		// fmt.Println("synapseToOutputCell", synapseToOutputCell)
		synapsesToEnd[synapseToOutputCell] = true
	}

	needSynapses := minSynapses - len(synapsesToEnd)
	if needSynapses > 0 {
		fmt.Println("GrowPathBetween adding end synapses ", needSynapses)
		hasWalked := len(alreadyWalked) > 0
		network.cellMux.Lock()
		fullEndCell := network.Cells[endCell]
		network.cellMux.Unlock()
		endHasDendrites := len(fullEndCell.DendriteSynapses) > 0
		// Two new synapse and one new cells will be added.
		// It will connect from the input network to a new cell to the end network.
		//
		// input cell or walked cell  ->  new synapse 1  ->  new cell  ->  new synapse 2 ->  end cell dendrite or end cell
		for i := 0; i < needSynapses; i++ {
			var startPathCell CellID
			var endPathCell CellID
			newInputSynapse := NewSynapse(network)
			newOutputSynapse := NewSynapse(network)
			newInputCell := NewCell(network)

			// TODO: perhaps make this deeper
			// TODO: add a method "create direct path between two cells"
			if hasWalked {
				// ordering of range map is random. select one.
				for cellID := range alreadyWalked {
					startPathCell = cellID
					break // yes break after one
				}
			} else {
				startPathCell = startCell
			}
			if endHasDendrites {
				// ordering of range map is random. select one.
				for synapseID := range fullEndCell.DendriteSynapses {
					network.synMux.Lock()
					endPathCell = network.Synapses[synapseID].FromNeuronAxon
					network.synMux.Unlock()
					break // yes break after one
				}
			} else {
				endPathCell = endCell
			}

			network.cellMux.Lock()
			network.Cells[startPathCell].addAxon(newInputSynapse.ID)
			network.cellMux.Unlock()
			newInputCell.addDendrite(newInputSynapse.ID)
			newInputCell.addAxon(newOutputSynapse.ID)
			network.cellMux.Lock()
			network.Cells[endPathCell].addDendrite(newOutputSynapse.ID)
			network.cellMux.Unlock()
		}
	}

	// Reinforce the path between expected input and output.
	// Since goodSynapses looks for cells that actually fired,
	// it may take more than one entire itration for this next
	// round of code to do anything. Because the new pathways
	// generated in the block above did not necessarily fire.
	goodSynapses := backwardTraceFirings(network, endCell, startCell)
	for synapseID := range goodSynapses {
		network.Synapses[synapseID].reinforce()
	}

	return synapsesToEnd, synapsesAdded
}

/*
GrowRandomNeurons will randomly add neurons with the default number of synapses to the network.
*/
func (network *Network) GrowRandomNeurons(neuronsToAdd, defaultNeuronSynapses int) {
	// fmt.Println("  adding neurons =", neuronsToAdd)
	// Now - all the new neurons are added first with no synapses. If synapses were added at
	// create time, the newer neurons would end up with far fewer connections to the following
	// newer neurons.
	var addedNeurons []*Cell
	for i := 0; i < neuronsToAdd; i++ {
		cell := NewCell(network)
		addedNeurons = append(addedNeurons, cell)
	}

	// Now we add the default number of synapses to our new neurons, with random other neurons.
	// Create the synapse, then choose a random cell from the network, then choose whether
	// this new cell will be a sender or receiver.
	for _, cell := range addedNeurons {
		for i := 0; i < defaultNeuronSynapses; {
			ix := network.RandomCellKey()
			otherCell := network.Cells[ix]
			if cell.ID == otherCell.ID {
				// try again
				continue
			}

			synapse := NewSynapse(network)

			if chooseIfSender() {
				synapse.FromNeuronAxon = cell.ID
				synapse.ToNeuronDendrite = otherCell.ID
				otherCell.DendriteSynapses[synapse.ID] = true
				cell.AxonSynapses[synapse.ID] = true
			} else {
				synapse.FromNeuronAxon = otherCell.ID
				synapse.ToNeuronDendrite = cell.ID
				otherCell.AxonSynapses[synapse.ID] = true
				cell.DendriteSynapses[synapse.ID] = true
			}
			i++
		}
	}
}

func chooseIfSender() bool {
	if randomIntBetween(0, 1) == 1 {
		return true
	}
	return false
}

/*
GrowRandomSynapses adds the specified number of synapses haphazardly to the network.
*/
func (network *Network) GrowRandomSynapses(synapsesToAdd int) {
	for i := 0; i < synapsesToAdd; {
		senderIx := network.RandomCellKey()
		receiverIx := network.RandomCellKey()
		sender := network.Cells[senderIx]
		receiver := network.Cells[receiverIx]
		// Thy cell shannot activate thyself
		if sender.ID == receiver.ID {
			continue
		}

		synapse := NewSynapse(network)
		synapse.ToNeuronDendrite = receiver.ID
		synapse.FromNeuronAxon = sender.ID
		sender.AxonSynapses[synapse.ID] = true
		receiver.DendriteSynapses[synapse.ID] = true
		i++
	}
}
