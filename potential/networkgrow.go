package potential

import (
	"fmt"
	"math"
	"sync"
)

// All methods on a network that relate to growing are here.

/*
Grow is a general growth that encompasses all growth methods.

It adds neurons, adds new synapses, prunes old neurons, and strengthens synapses that have
fired a lot.
*/
func (network *Network) Grow(neuronsToAdd, defaultNeuronSynapses, synapsesToAdd int) {
	fmt.Println("Grow session start")
	network.Prune()

	network.GrowRandomNeurons(neuronsToAdd, defaultNeuronSynapses)

	network.GrowRandomSynapses(synapsesToAdd)
	fmt.Println("  Grow session end, synapses=", len(network.Synapses), "cells=", len(network.Cells))
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

TODO: finish GrowPathBetween
*/
func (network *Network) GrowPathBetween(startCell, endCell CellID, minSynapses int) (synapsesToEnd map[SynapseID]bool, synapsesAdded map[SynapseID]bool) {
	// these are the synapses we found that are on the path from the startCell,
	// and attach directly to an endCell at the dendrite
	synapsesToEnd = make(map[SynapseID]bool)
	// any new synapses we create if there are not enough in the network that attach
	// to the end cell
	synapsesAdded = make(map[SynapseID]bool)

	avgSynPerCell := float64(len(network.Synapses) / len(network.Cells))
	// semi-hardcoded number of max hops. this was arbitrary.
	maxHops := int(math.Max(math.Min(avgSynPerCell, 50.0), 20))

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
					fmt.Println("warn: synapse does not exist", axonSynapseID,
						"from cell=", cellID)
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
			synapse := NewSynapse()
			network.Synapses[synapse.ID] = synapse
			synapsesAdded[synapse.ID] = true
			synapse.Network = network
			// somewhat arbitrarily decided to set the synapses to the highest value allowed
			synapse.Millivolts = int8(synapseMax)

			synapse.FromNeuronAxon = lastCellID
			network.Cells[lastCellID].AxonSynapses[synapse.ID] = true

			synapse.ToNeuronDendrite = endCell
			network.Cells[endCell].DendriteSynapses[synapse.ID] = true
		}
	}
	return synapsesToEnd, synapsesAdded
}

/*
Prune traverses the network looking for neurons to degrade.

This degrades synapses that didn't fire during a certain round of testing.

Once they have been degraded to no net effect (0 millivolts), they should be removed.

Additionally, it checks when the cells that connect to a synapse have no more synapses,
and removes those cells if they have none.
*/
func (network *Network) Prune() {
	// Next we move the less used synapses toward zero, because doing this later would prune the
	// brand new synapses. This is a good time to apply the learning rate to synapses
	// which were activated, too.
	var synapsesToRemove []*Synapse
	// fmt.Println("  processing learning on cells, total=", len(network.Cells))
	for _, cell := range network.Cells {
		for synapseID := range cell.DendriteSynapses { // could also be axons, but, meh.
			synapse, exists := network.Synapses[synapseID]
			if !exists {
				fmt.Println("warn: synapse attempted to be accesed but it was already removed", synapseID)
				continue
			}
			isPositive := synapse.Millivolts >= 0

			if synapse.ActivationHistory >= network.SynapseMinFireThreshold {
				// it was activated enough, so we bump it away from zero.
				// needs cleanup refactoring.
				if isPositive {
					if int16(synapse.Millivolts)-int16(network.SynapseLearnRate) > -127 {
						synapse.Millivolts -= network.SynapseLearnRate // do not overflow int8
					} else {
						synapse.Millivolts = -128
					}
				} else {
					if int16(synapse.Millivolts)+int16(network.SynapseLearnRate) < 126 {
						synapse.Millivolts += network.SynapseLearnRate
					} else {
						synapse.Millivolts = 127
					}
				}
			} else {
				// It reached "no effect" of 0 millivolts last round. Then it didn't fire
				// this round. Remove this synapse.
				if synapse.Millivolts == 0 {
					synapsesToRemove = append(synapsesToRemove, synapse)
				}

				// did not meet minimum fire threshold, so punish it by moving toward zero
				if isPositive {
					synapse.Millivolts -= network.SynapseLearnRate
				} else {
					synapse.Millivolts += network.SynapseLearnRate
				}

				// Next time, if it did not fire, and it is zero, it will get pruned.
			}

			// reset the activation history until the next Grow cycle.
			synapse.ActivationHistory = 0
		}
	}
	// fmt.Println("  done")

	fmt.Println("  synapses to remove=", len(synapsesToRemove))
	// Actually pruning synapses is done after the previous loop because it can
	// trigger removal of Cells, which can subsequently mess up the range operation
	// happening over the same array of cells.
	for _, synapse := range synapsesToRemove {
		network.PruneSynapse(synapse.ID)
	}
	fmt.Println("  done pruning")
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
		cell := NewCell()
		network.Cells[cell.ID] = cell
		cell.Network = network
		addedNeurons = append(addedNeurons, cell)
	}
	// fmt.Println("  done")

	// fmt.Println("  adding default synapses to new neurons", defaultNeuronSynapses)

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

			synapse := NewSynapse()
			network.Synapses[synapse.ID] = synapse
			synapse.Network = network

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
			// fmt.Println("created synapse", synapse)
			i++
		}
	}
	// fmt.Println("  GrowRandomNeurons done")
}

/*
GrowRandomSynapses adds the specified number of synapses haphazardly to the network.
*/
func (network *Network) GrowRandomSynapses(synapsesToAdd int) {
	// fmt.Println("  adding synapses to whole network", synapsesToAdd)
	// Then we randomly add synapses between neurons to the whole network, including the
	// newest neurons.
	for i := 0; i < synapsesToAdd; {
		senderIx := network.RandomCellKey()
		receiverIx := network.RandomCellKey()
		sender := network.Cells[senderIx]
		receiver := network.Cells[receiverIx]
		// Thy cell shannot activate thyself
		if sender.ID == receiver.ID {
			continue
		}

		synapse := NewSynapse()
		network.Synapses[synapse.ID] = synapse
		synapse.Network = network
		synapse.ToNeuronDendrite = receiver.ID
		synapse.FromNeuronAxon = sender.ID
		sender.AxonSynapses[synapse.ID] = true
		receiver.DendriteSynapses[synapse.ID] = true
		// fmt.Println("created synapse", synapse)
		i++
	}
	// fmt.Println("  GrowRandomSynapses done")

}

/*
PruneSynapse removes a synapse and all references to it.

A synapse exists between two neurons. We get both neurons and remove the synapse
from its list.

If either of those neurons no longer has any synapses itself, kill off that neuron cell.
*/
func (network *Network) PruneSynapse(synapseID SynapseID) {
	synapse, ok := network.Synapses[synapseID]
	if !ok {
		fmt.Println("warn: attempt to remove synapse that is not in network", synapseID)
		return
	}

	// See if either cell (to/from) should be pruned, also.
	// Technically this can result in a cell being the end of a dead pathway, or not receiving
	// any input. But that is something to revisit. It is likely these cells would eventually
	// build up more synapses via the grow process.

	if dendriteCell, exists := network.Cells[synapse.FromNeuronAxon]; exists {
		delete(dendriteCell.AxonSynapses, synapse.ID)
		receiverCellHasNoSynapses := len(dendriteCell.AxonSynapses) == 0 && len(dendriteCell.DendriteSynapses) == 0
		if !dendriteCell.Immortal && receiverCellHasNoSynapses {
			delete(network.Cells, dendriteCell.ID)
		}
	}

	if axonCell, exists := network.Cells[synapse.ToNeuronDendrite]; exists {
		delete(axonCell.DendriteSynapses, synapse.ID)

		senderCellHasNoSynapses := len(axonCell.AxonSynapses) == 0 && len(axonCell.DendriteSynapses) == 0
		if !axonCell.Immortal && senderCellHasNoSynapses {
			delete(network.Cells, axonCell.ID)
		}
	}

	delete(network.Synapses, synapse.ID)
	// this synapse is now dead
}
