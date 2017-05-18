package potential

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewCell(t *testing.T) {
	t.Run("calling NewCell() adds it to the network", func(t *testing.T) {
		network := NewNetwork()
		cell := NewCell(network)
		// network.Cells[cell.ID] = cell
		// cell.Network = network
		assert.Equal(t, 1, len(network.Cells))
		ok := network.cellExists(cell.ID)
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
		assert.Equal(t, 1, len(network.Cells))
		network.PruneCell(cell.ID)
		assert.Equal(t, 0, len(network.Cells))
	})
	t.Run("does not remove an immortal cell from the network", func(t *testing.T) {
		network := NewNetwork()
		cell := NewCell(network)
		cell.Immortal = true
		assert.Equal(t, 1, len(network.Cells))
		network.PruneCell(cell.ID)
		assert.Equal(t, 1, len(network.Cells))
	})
	t.Skip("panics when there are synapses still on the cell", func(t *testing.T) {
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
