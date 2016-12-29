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
	var lastCellID CellID
	hops := 0
	var wg sync.WaitGroup
	ch := make(chan SynapseID)

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
		lastCellID = cellID
		hops++
		totalSynapsesFound := len(synapsesToEnd)
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
				// also walk the axons of this cell, and pipe any values downstream.
				walk(receiverCellID)
			}
			wg.Done()
		}()
	}

	// receive the connectd synapses.
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
		// Add cells and synapases from the last cell we saw to the output.
		// Whatever the last cell at `maxHops` from the input was, will get
		// `needSynapses` more synapses directly to the `endCell`.
		for i := 0; i < needSynapses; i++ {
			synapse := NewSynapse(network)
			synapsesAdded[synapse.ID] = true
			// somewhat arbitrarily decided to set the synapses to the highest value allowed
			synapse.Millivolts = synapseLearnRate

			synapse.FromNeuronAxon = lastCellID
			network.Cells[lastCellID].AxonSynapses[synapse.ID] = true

			synapse.ToNeuronDendrite = endCell
			network.Cells[endCell].DendriteSynapses[synapse.ID] = true
		}
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
