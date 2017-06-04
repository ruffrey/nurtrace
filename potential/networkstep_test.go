package potential

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NetworkStep(t *testing.T) {
	t.Run("postsynaptic cells of flagged synapses get flagged as activating", func(t *testing.T) {
		network := NewNetwork()
		cell := NewCell(network)
		receiverCellA := NewCell(network)
		receiverCellB := NewCell(network)

		// setup two synapses, from cell to each receiver
		network.linkCells(cell.ID, receiverCellA.ID)
		network.linkCells(cell.ID, receiverCellB.ID)

		network.GetSyn(0).fireNextRound = true
		network.GetSyn(1).fireNextRound = true

		// some pretesting
		assert.Equal(t, false, receiverCellA.activating)
		assert.Equal(t, false, receiverCellB.activating)

		// actual test
		hasMore := network.Step()

		assert.Equal(t, false, hasMore,
			"should not be another round to fire")

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
	})
	t.Skip("when firing one round results in firing the next round it returns true")
	t.Skip("when firing one round results in no new firings it returns false")
}
