package potential

import "fmt"

/*
Step fires the next round of cells.

The current array of cells to be fired get removed, and the cells on their
synapses are found and replaced in the firing list.

When a cell fires, we stop its activation for the next step, much like a real
neuron will go through a refractory period after it fires.
*/
func (network *Network) Step() (hasMore bool) {
	nextCellResets := make(map[CellID]bool)

	for synapseID := range network.nextSynapsesToActivate {
		syn, exists := network.Synapses[synapseID]
		if !exists {
			fmt.Println("error: synapse cannot be activated because it does not exist")
			continue
		}

		didFireCell, err := syn.Activate()
		if err != nil {
			fmt.Println("error: failed activating synapse", synapseID, err)
			continue
		}
		if didFireCell {
			nextCellResets[syn.ToNeuronDendrite] = true
		}
	}

	for cellID := range network.resetCellsOnNextStep {
		cell, exists := network.Cells[cellID]
		if !exists {
			fmt.Println("error: cell cannot be reset because it does not exist")
			continue
		}
		cell.Voltage = cellRestingVoltage
		cell.activating = false
	}

	hasMore = len(nextCellResets) > 0

	network.nextSynapsesToActivate = make(map[SynapseID]bool)
	network.resetCellsOnNextStep = nextCellResets

	return hasMore
}

/*
AddSynapseToNextStep provides a reusable method for having a synapse get activated on the
next step.
*/
func (network *Network) AddSynapseToNextStep(id SynapseID) {
	_, exists := network.Synapses[id]
	if !exists {
		fmt.Println("error: cannot add synapse", id, "to activation list because it does not exist")
		return
	}
	network.nextSynapsesToActivate[id] = true
}
