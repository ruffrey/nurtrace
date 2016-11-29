package potential

import "fmt"

// All methods on a network that relate to growing are here.

/*
Grow is a general growth that encompasses all growth methods.

It adds neurons, adds new synapses, prunes old neurons, and strengthens synapses that have
fired a lot.
*/
func (network *Network) Grow(neuronsToAdd, defaultNeuronSynapses, synapsesToAdd int) {
	network.Prune()

	network.GrowRandomNeurons(neuronsToAdd, defaultNeuronSynapses)

	network.GrowRandomSynapses(synapsesToAdd)
	// fmt.Println("  Grow session end")
}

/*
GrowPathBetween will make a path between two neurons.

This is a little complicated and probably should only be used sparingly.

The desired goal is for there to be a path from startCell to endCell.

We start at both cells, working forward through the axons of startCell and its
connections, and backward through the dendrites of endCell and its connections.

At each layer, we check to see if any of the start axons are connected to the
end dendrites. If there is no connection, we follow all those axons and dendrites.

After `maxHops`, if there are not minSynapses, we create synapses at that layer.

TODO: finish GrowPathBetween
*/
func (network *Network) GrowPathBetween(startCell, endCell CellID, maxHops, minSynapses int) {

}

/*
Prune traverses the network looking for neurons to degrade.

This degrades synapses that didn't fire during a certain round of testing.

Once they have been degraded to no net effect (0 millivolts), they should be removed.

Additionally, it checks when the cells that connect to a synapse have no more synapses,
and removes those cells if they have none.
*/
func (network *Network) Prune() {
	// fmt.Println("Grow session start")
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
		network.PruneSynapse(synapse)
	}
	// fmt.Println("  done")
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
	for _, cell := range addedNeurons {
		for i := 0; i < defaultNeuronSynapses; {
			synapse := NewSynapse()
			network.Synapses[synapse.ID] = synapse
			synapse.Network = network

			ix := network.RandomCellKey()
			otherCell := network.Cells[ix]
			if cell.ID == otherCell.ID {
				// try again
				continue
			}
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
	// fmt.Println("  done")
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
	// fmt.Println("  done")

}

/*
PruneSynapse removes a synapse and all references to it.

A synapse exists between two neurons. We get both neurons and remove the synapse
from its list.

If either of those neurons no longer has any synapses itself, kill off that neuron cell.
*/
func (network *Network) PruneSynapse(synapse *Synapse) {
	dendriteCell := network.Cells[synapse.FromNeuronAxon]
	axonCell := network.Cells[synapse.ToNeuronDendrite]

	delete(dendriteCell.AxonSynapses, synapse.ID)
	delete(axonCell.DendriteSynapses, synapse.ID)

	// See if either cell (to/from) should be pruned, also.
	// Technically this can result in a cell being the end of a dead pathway, or not receiving
	// any input. But that is something to revisit. It is likely these cells would eventually
	// build up more synapses via the grow process.
	receiverCellHasNoSynapses := len(dendriteCell.AxonSynapses) == 0 && len(dendriteCell.DendriteSynapses) == 0
	if receiverCellHasNoSynapses {
		delete(network.Cells, dendriteCell.ID)
	}
	senderCellHasNoSynapses := len(axonCell.AxonSynapses) == 0 && len(axonCell.DendriteSynapses) == 0
	if senderCellHasNoSynapses {
		delete(network.Cells, axonCell.ID)
	}

	delete(network.Synapses, synapse.ID)
	// this synapse is now dead
}
