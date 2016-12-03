package potential

import "fmt"

/*
Prune traverses the network looking for neurons to degrade or reinforce.

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

			// when applying voltages, we must be careful to not overflow the int8 size
			if synapse.ActivationHistory >= network.SynapseMinFireThreshold {
				// it was activated enough, so we bump it away from zero.
				// needs cleanup refactoring.
				if isPositive {
					newMV := synapse.Millivolts + network.SynapseLearnRate
					if newMV > network.actualSynapseMax {
						synapse.Millivolts = network.actualSynapseMax
					} else {
						synapse.Millivolts = newMV
					}
				} else {
					newMV := synapse.Millivolts - network.SynapseLearnRate
					if newMV < network.actualSynapseMin {
						synapse.Millivolts = network.actualSynapseMin
					} else {
						synapse.Millivolts = newMV
					}
				}
			} else if synapse.ActivationHistory > 0 {
				// did not meet minimum fire threshold, so punish it by moving toward zero
				if isPositive {
					newMV := synapse.Millivolts + network.SynapseLearnRate
					if newMV > network.actualSynapseMax {
						synapse.Millivolts = network.actualSynapseMax
					} else {
						synapse.Millivolts = newMV
					}
				} else {
					newMV := synapse.Millivolts - network.SynapseLearnRate
					if newMV < network.actualSynapseMin {
						synapse.Millivolts = network.actualSynapseMax
					} else {
						synapse.Millivolts = newMV
					}
				}
			} else {
				// it never fired during the training session, so it should be removed.
				synapsesToRemove = append(synapsesToRemove, synapse)
			}

			// Reset the activation history until the next Prune cycle.
			// This is not done in ResetForTraining because it is used across trainings.
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
PruneSynapse removes a synapse and all references to it.

A synapse exists between two neurons. We get both neurons and remove the synapse
from its list.

If either of those neurons no longer has any synapses itself, kill off that neuron cell.

Unless the neuron is immortal, then just remove the synapse.
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
	// build up more synapses via the grow process, or never fire and be pruned later.

	if c, exists := network.Cells[synapse.FromNeuronAxon]; exists {
		delete(c.AxonSynapses, synapse.ID)
		cellHasNoSynapses := len(c.AxonSynapses) == 0 && len(c.DendriteSynapses) == 0
		if cellHasNoSynapses {
			network.PruneCell(c.ID)
		}
	}

	if c, exists := network.Cells[synapse.ToNeuronDendrite]; exists {
		delete(c.DendriteSynapses, synapse.ID)

		cellHasNoSynapses := len(c.AxonSynapses) == 0 && len(c.DendriteSynapses) == 0
		if cellHasNoSynapses {
			network.PruneCell(c.ID)
		}
	}

	delete(network.Synapses, synapse.ID)
	// this synapse is now dead
}

/*
PruneCell removes a cell and its synapses. It is independent of PruneSynapse.
*/
func (network *Network) PruneCell(cellID CellID) {
	cell, ok := network.Cells[cellID]
	if !ok {
		fmt.Println("warn: attempt to prune cell which does not exist", cellID)
		return
	}
	if cell.Immortal {
		return
	}
	for synapseID := range cell.DendriteSynapses {
		network.PruneSynapse(synapseID)
	}
	for synapseID := range cell.AxonSynapses {
		network.PruneSynapse(synapseID)
	}
}
