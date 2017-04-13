package potential

import (
	"fmt"
	"math"

	"github.com/ruffrey/nurtrace/laws"
)

/*
firingGroup is temporarily used to store a group of synapses that
fired and which cell they fired onto. It is useful so we can
look back and boost the synapses which resulted in the cell
firing due to sufficient voltage increase.
*/
type firingGroup struct {
	synapses map[SynapseID]bool
	voltage  int
}

func newFiringGroup(forCell *Cell) *firingGroup {
	return &firingGroup{
		synapses: make(map[SynapseID]bool),
		voltage:  int(forCell.Voltage),
	}
}

/*
Step fires the next round of synapses.

During this process, we tally up the voltage that would be applied to
each cell. We do this in one step so it is more predictable, and it
"allows" all synapses to act on the cell at once, rather than whatever
synapse happened to fire first (random based on cpu factors). It is
not clear that this is the best solution, but applying synapse voltages
immediately did not appear to lead to a usable network.

When a cell fires, we stop its activation for the next step, much like a real
neuron will go through a refractory period after it fires.
*/
func (network *Network) Step() (hasMore bool) {
	if network.Disabled {
		return false
	}

	nextCellResets := make(map[CellID]bool) // these cells got fired
	voltageTallies := make(map[CellID]*firingGroup)

	// tally up all the synapse voltage results they will have on the cells
	for synapseID := range network.nextSynapsesToActivate {
		syn, exists := network.Synapses[synapseID]
		if !exists {
			fmt.Println("error: synapse cannot be activated because it does not exist")
			continue
		}

		cellReceivingVoltage := network.getCell(syn.ToNeuronDendrite)
		if cellReceivingVoltage.activating { // do not fire cells in refractory period
			continue
		}
		if _, seen := voltageTallies[cellReceivingVoltage.ID]; !seen {
			voltageTallies[cellReceivingVoltage.ID] = newFiringGroup(cellReceivingVoltage)
		}
		fg := voltageTallies[cellReceivingVoltage.ID]
		fg.voltage += int(syn.Millivolts)
		// save the synapse for later so it can be boosted if the cell fires
		fg.synapses[syn.ID] = true
	}
	// Reset this list of synapses now that we activated them. The next loop starts
	// adding more for the next round.
	network.nextSynapsesToActivate = make(map[SynapseID]bool)

	// see if any cells fired
	for cellID, fg := range voltageTallies {
		cell := network.getCell(cellID)
		if fg.voltage < laws.CellFireVoltageThreshold {
			// prevent out of bounds voltage
			cell.Voltage = int16(math.Max(float64(fg.voltage), float64(laws.ActualSynapseMin)))
			continue
		}

		// It should fire.

		cell.FireActionPotential()
		nextCellResets[cellID] = true

		// Reward the synapses that were involved in this cell firing.
		for synapseID := range fg.synapses {
			fromSynapse := network.getSyn(synapseID)
			fromSynapse.reinforce()
		}
	}

	// for the cells from the last step, make them fire-able again
	for cellID := range network.resetCellsOnNextStep {
		cell, exists := network.Cells[cellID]
		if !exists {
			fmt.Println("error: cell cannot be reset because it does not exist")
			continue
		}
		cell.Voltage = laws.CellRestingVoltage
		cell.activating = false
	}

	hasMore = len(nextCellResets) > 0

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
