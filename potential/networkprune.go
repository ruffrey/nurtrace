package potential

import (
	"fmt"
	"math"

	"github.com/y0ssar1an/q"
)

/*
Prune traverses the network looking for neurons to degrade or reinforce.

This degrades synapses that didn't fire during a certain round of testing.

Once they have been degraded to no net effect (0 millivolts), they should be removed.

Additionally, it checks when the cells that connect to a synapse have no more synapses,
and removes those cells if they have none.
*/
func (network *Network) Prune() {
	network.nextSynapsesToActivate = make(map[SynapseID]bool)
	network.resetCellsOnNextStep = make(map[CellID]bool)
	if ok, report := CheckIntegrity(network); !ok {
		q.Q("Prune: network has no integrity BEFORE pruning", report)
		panic("no integrity")
	}
	// Next we move the less used synapses toward zero, because doing this later would prune the
	// brand new synapses. This is a good time to apply the learning rate to synapses
	// which were activated, too.
	synapsesToRemove := make(map[SynapseID]bool)
	// fmt.Println("  processing learning on cells, total=", len(network.Cells))
	for _, cell := range network.Cells {
		for synapseID := range cell.AxonSynapses { // could also be axons, but, meh.
			synapse, exists := network.Synapses[synapseID]
			if !exists {
				fmt.Println("warn: cannot evaluate synapse", synapseID,
					"- missing from cell dendrites", cell.ID)
				continue
			}
			isPositive := synapse.Millivolts >= 0

			// when applying voltages, we must be careful to not overflow the int8 size
			if synapse.ActivationHistory >= defaultSynapseMinFireThreshold {
				// it was activated enough, so we bump it away from zero.
				// needs cleanup refactoring.
				if isPositive {
					newMV := synapse.Millivolts + synapseLearnRate
					if newMV > actualSynapseMax {
						synapse.Millivolts = actualSynapseMax
					} else {
						synapse.Millivolts = newMV
					}
				} else {
					newMV := synapse.Millivolts - synapseLearnRate
					if newMV < actualSynapseMin {
						synapse.Millivolts = actualSynapseMin
					} else {
						synapse.Millivolts = newMV
					}
				}
			} else if synapse.ActivationHistory > 0 {
				// did not meet minimum fire threshold, so punish it by moving toward zero
				distanceToZero := math.Abs(0 - float64(synapse.Millivolts))
				halfDistance := int8(math.Ceil(distanceToZero / 2))
				if halfDistance < 3 {
					synapse.Millivolts = 0
				} else if isPositive {
					synapse.Millivolts -= halfDistance
				} else {
					synapse.Millivolts += halfDistance
				}
			} else {
				// it never fired during the training session, so it should be removed.
				synapsesToRemove[synapse.ID] = true
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
	for synapseID := range synapsesToRemove {
		network.PruneSynapse(synapseID)
	}
	fmt.Println("  done pruning")

	if ok, report := CheckIntegrity(network); !ok {
		fmt.Println("Prune: network has no integrity AFTER pruning")
		report.Print()
		fmt.Println("intended to remove synapses:", synapsesToRemove)
		for cellID, synapseID := range report.cellHasMissingAxonSynapse {
			fmt.Println("  cellHasMissingAxonSynapse", cellID, network.Cells[cellID])
			fmt.Println("  cellHasMissingAxonSynapse", synapseID, network.Synapses[synapseID])
		}
		for cellID, synapseID := range report.cellHasMissingDendriteSynapse {
			fmt.Println("  cellHasMissingDendriteSynapse", cellID, network.Cells[cellID])
			fmt.Println("  cellHasMissingDendriteSynapse", synapseID, network.Synapses[synapseID])
		}
		panic("no integrity")
	}
}

/*
PruneSynapse removes a synapse and all references to it.

A synapse exists between two neurons. We get both neurons and remove the synapse
from its list.

If either of those neurons no longer has any synapses itself, kill off that neuron cell.

Unless the neuron is immortal, then just remove the synapse.
*/
func (network *Network) PruneSynapse(synapseID SynapseID) {
	// fmt.Println("remove synapse=", synapseID)
	var synapse *Synapse

	synapse, ok := network.Synapses[synapseID]
	if !ok {
		fmt.Println("warn: attempt to remove synapse that is not in network", synapseID)
		return
	}

	// See if either cell (to/from) should be pruned, also.
	// Technically this can result in a cell being the end of a dead pathway, or not receiving
	// any input. But that is something to revisit. It is likely these cells would eventually
	// build up more synapses via the grow process, or never fire and be pruned later.

	network.removeSynapseFromCell(synapseID, synapse.FromNeuronAxon, true)
	network.removeSynapseFromCell(synapseID, synapse.ToNeuronDendrite, false)

	delete(network.Synapses, synapse.ID)
	// this synapse is now dead
}

func (network *Network) removeSynapseFromCell(s SynapseID, c CellID, isAxon bool) {
	cell, exists := network.Cells[c]
	if exists {
		if isAxon {
			delete(cell.AxonSynapses, s)
		} else {
			delete(cell.DendriteSynapses, s)
		}
	} else {
		fmt.Println("warn: cannot prune synapse", s, "(isAxon=", isAxon, ") from cell",
			c, " which does not exist")
	}

	// after removing the synapses, see if this cell can be removed.
	cellHasNoSynapses := len(cell.AxonSynapses) == 0 && len(cell.DendriteSynapses) == 0
	if cellHasNoSynapses {
		network.PruneCell(c)
	}
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

	// with good code, the following should not be necessary.
	for synapseID := range cell.DendriteSynapses {
		fmt.Println("warn: should not need to prune dendrite synapse from cell")
		network.PruneSynapse(synapseID)
	}
	for synapseID := range cell.AxonSynapses {
		fmt.Println("warn: should not need to prune axon synapse from cell")
		network.PruneSynapse(synapseID)
	}

	// Do this after removing synapses, because otherwise we can end up with
	// orphan synapses.
	if cell.Immortal {
		return
	}

	delete(network.Cells, cellID)
}
