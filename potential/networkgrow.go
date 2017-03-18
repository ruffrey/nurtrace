package potential

import (
	"bleh/laws"
	"fmt"
	"sync"
)

// All methods on a network that relate to growing are here.

/*
Grow is a general growth that encompasses all growth methods.

It adds neurons, adds new synapses, prunes old neurons, and strengthens synapses that have
fired a lot.
*/
func (network *Network) Grow(neuronsToAdd, synapsesPerNewNeuron, synapsesToAdd int) {
	network.GrowRandomNeurons(neuronsToAdd, synapsesPerNewNeuron)
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
	var walk func(cellID CellID)
	walk = func(cellID CellID) {
		mux.Lock()
		hops++
		totalSynapsesFound := len(synapsesToEnd)
		if hops >= maxHops || totalSynapsesFound >= minSynapses {
			mux.Unlock()
			return
		}
		if _, already := alreadyWalked[cellID]; already {
			mux.Unlock()
			return
		}
		alreadyWalked[cellID] = true
		mux.Unlock()

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
		hasWalked := len(alreadyWalked) > 0
		fmt.Println("start=", startCell, "end=", endCell, "minSynapses", minSynapses, "needSynapses", needSynapses, "alreadyWalked", len(alreadyWalked), "maxHops", network.maxHops)
		// Two new synapse and one new cell will be added.
		// It will connect from the input network to a new cell to the end network.
		//
		// input cell or walked cell  ->  new synapse 1  ->  new cell  ->  new synapse 2 ->  end cell dendrite or end cell
		for i := 0; i < needSynapses; i++ {
			var startPathCell CellID

			if hasWalked {
				// ordering of range map is random. select one.
				startPathCell = randCellFromMap(alreadyWalked)
			} else {
				startPathCell = startCell
			}

			newLinkingSynapse := network.linkCells(startPathCell, endCell)

			synapsesAdded[newLinkingSynapse.ID] = true

			newLinkingSynapse.Millivolts = laws.DefaultNewGrownPathSynapse
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
func (network *Network) GrowRandomNeurons(neuronsToAdd, synapsesPerNeuron int) {
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
		for i := 0; i < synapsesPerNeuron; {
			ix := network.RandomCellKey()
			otherCell := network.Cells[ix]
			if cell.ID == otherCell.ID {
				// try again
				continue
			}

			if chooseIfSender() {
				network.linkCells(cell.ID, otherCell.ID)
			} else {
				network.linkCells(otherCell.ID, cell.ID)
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

		network.linkCells(sender.ID, receiver.ID)
		i++
	}
}
