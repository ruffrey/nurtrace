package potential

import (
	"testing"

	"github.com/ruffrey/nurtrace/laws"
	"github.com/stretchr/testify/assert"
)

func Test_NetworkStep(t *testing.T) {
	t.Run("postsynaptic cells of flagged synapses get flagged as activating", func(t *testing.T) {
		network := NewNetwork()
		cell := NewCell(network)
		receiverCellA := NewCell(network)
		receiverCellB := NewCell(network)
		receiverCellA.Voltage = 0
		receiverCellB.Voltage = 0

		// setup two synapses, from cell to each receiver
		network.linkCells(cell.ID, receiverCellA.ID)
		network.linkCells(cell.ID, receiverCellB.ID)

		network.GetSyn(0).fireNextRound = true
		network.GetSyn(1).fireNextRound = true
		// enough to fire next cell
		network.GetSyn(0).Millivolts = int16(laws.CellFireVoltageThreshold)
		network.GetSyn(1).Millivolts = int16(laws.CellFireVoltageThreshold)

		// some pretesting
		assert.Equal(t, false, receiverCellA.activating)
		assert.Equal(t, false, receiverCellB.activating)

		// actual test
		hasMore := network.Step()

		assert.Equal(t, true, hasMore,
			"should have resets on next Step")

		// Now the cells are activating
		assert.Equal(t, true, receiverCellA.activating,
			"cell is not flagged as activating after Step")
		assert.Equal(t, true, receiverCellB.activating,
			"cell is not flagged as activating after Step")

		// and the synapses have been reset
		assert.Equal(t, false, network.GetSyn(0).fireNextRound,
			"synapse has not been reset for firing after Step")
		assert.Equal(t, false, network.GetSyn(1).fireNextRound,
			"synapse has not been reset for firing after Step")

		// next Step should exhaust everything as there are no more
		// synapses to fire
		hasMore = network.Step()
		assert.Equal(t, false, hasMore,
			"Step should not yield more cells to reset next round")
	})
	t.Run("when firing one round results in no next firings it returns false", func(r *testing.T) {
		network := NewNetwork()
		cell := NewCell(network)
		receiverCellA := NewCell(network)
		receiverCellB := NewCell(network)
		receiverCellA.Voltage = 0
		receiverCellB.Voltage = 0

		// setup two synapses, from cell to each receiver
		network.linkCells(cell.ID, receiverCellA.ID)
		network.linkCells(cell.ID, receiverCellB.ID)

		network.GetSyn(0).fireNextRound = false
		network.GetSyn(1).fireNextRound = false
		// enough to fire next cell
		network.GetSyn(0).Millivolts = int16(laws.CellFireVoltageThreshold)
		network.GetSyn(1).Millivolts = int16(laws.CellFireVoltageThreshold)

		// some pretesting
		assert.Equal(t, false, receiverCellA.activating)
		assert.Equal(t, false, receiverCellB.activating)

		// actual test
		hasMore := network.Step()

		assert.Equal(t, false, hasMore,
			"should not have resets on next Step")
	})
}
