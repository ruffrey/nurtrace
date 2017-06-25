package potential

import (
	"testing"

	"github.com/ruffrey/nurtrace/laws"
	"github.com/stretchr/testify/assert"
)

func Test_NewCell(t *testing.T) {
	t.Run("calling NewCell() adds it to the network", func(t *testing.T) {
		network := NewNetwork()
		cell := NewCell(network)
		// network.Cells[cell.ID] = cell
		// cell.Network = network
		assert.Equal(t, 1, len(network.Cells))
		ok := network.CellExists(cell.ID)
		if !ok {
			panic("NewCell() did not add cell to the network")
		}
		netCell := network.Cells[cell.ID]
		assert.Equal(t, cell.ID, netCell.ID, "NewCell() added cell to network but ids do not match")
	})
	t.Run("creates proper index / ID", func(t *testing.T) {
		network := NewNetwork()

		network.GrowRandomNeurons(10, 4)

		cell := NewCell(network)
		assert.Equal(t, cell.ID, network.Cells[cell.ID].ID)

		for index, c := range network.Cells {
			assert.Equal(t, int(c.ID), index)
		}
	})
}

func Test_CellFireAP(t *testing.T) {
	t.Run("firing cell results in its synapses having the fire-next flag set", func(t *testing.T) {
		network := NewNetwork()
		cell := NewCell(network)
		receiverCell := NewCell(network)

		// two synapses from cell to receiver
		network.linkCells(cell.ID, receiverCell.ID)
		network.linkCells(cell.ID, receiverCell.ID)

		// pretest
		assert.Equal(t, false, network.GetSyn(0).fireNextRound,
			"bad initial fire state of synapse")
		assert.Equal(t, false, network.GetSyn(1).fireNextRound,
			"bad initial fire state of synapse")

		// test
		network.GetCell(0).FireActionPotential()
		assert.Equal(t, true, network.GetSyn(0).fireNextRound,
			"synapse should be flagged for firing next round")
		assert.Equal(t, true, network.GetSyn(1).fireNextRound,
			"synapse should be flagged for firing next round")
	})
}

func Test_CellStringer(t *testing.T) {
	t.Run("String works without crashing", func(t *testing.T) {
		network := NewNetwork()
		cell := NewCell(network)
		synapse := NewSynapse(network)
		cell.AxonSynapses[synapse.ID] = true
		cell.DendriteSynapses[synapse.ID] = true
		s := cell.String()
		assert.NotEmpty(t, s)
	})
}

func Test_PruneCell(t *testing.T) {
	t.Run("removes a non-Immortal cell from the network", func(t *testing.T) {
		network := NewNetwork()
		cell := NewCell(network)
		cell.Immortal = false
		assert.Equal(t, true, network.CellExists(cell.ID))
		network.PruneCell(cell.ID)
		assert.Equal(t, false, network.CellExists(cell.ID))
	})
	t.Run("does not remove an immortal cell from the network", func(t *testing.T) {
		network := NewNetwork()
		cell := NewCell(network)
		cell.Immortal = true
		assert.Equal(t, 1, len(network.Cells))
		network.PruneCell(cell.ID)
		assert.Equal(t, 1, len(network.Cells))
	})
	t.Run("panics when there are synapses still on the cell", func(t *testing.T) {
		network := NewNetwork()
		cell1 := NewCell(network)
		cell1.DendriteSynapses[556] = true
		cell2 := NewCell(network)
		cell2.AxonSynapses[213] = true

		assert.Panics(t, func() {
			network.PruneCell(cell1.ID)
		})
		assert.Panics(t, func() {
			network.PruneCell(cell2.ID)
		})
	})
}

func Test_CellTowardResting(t *testing.T) {
	t.Run("when voltage is already resting it keeps it", func(t *testing.T) {
		network := NewNetwork()
		cell := NewCell(network)

		cell.Voltage = laws.CellRestingVoltage
		cell.towardResting()
		assert.Equal(t, laws.CellRestingVoltage, cell.Voltage)
	})
	t.Run("when voltage is greater than resting, it moves toward zero", func(t *testing.T) {
		network := NewNetwork()
		cell := NewCell(network)

		// first test
		cell.Voltage = laws.CellRestingVoltage + 50
		cell.towardResting()
		assert.Equal(t, int16(8), cell.Voltage)

		// another
		cell.Voltage = laws.CellRestingVoltage + 9
		cell.towardResting()
		assert.Equal(t, int16(-23), cell.Voltage)
	})
	t.Run("when voltage is less than resting, it moves toward zero", func(t *testing.T) {
		network := NewNetwork()
		cell := NewCell(network)

		// first test
		cell.Voltage = laws.CellRestingVoltage - 22
		cell.towardResting()
		assert.Equal(t, int16(-47), cell.Voltage)

		// another
		cell.Voltage = laws.CellRestingVoltage - 67
		cell.towardResting()
		assert.Equal(t, int16(-81), cell.Voltage)
	})
}
