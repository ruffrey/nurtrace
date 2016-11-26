package potential

import "testing"

func Test_NewNetwork(t *testing.T) {
	original := NewNetwork()

	t.Run("Cell map is initialized and can add cell immediately", func(t *testing.T) {
		cell := NewCell(&original)
		original.Cells[cell.ID] = &cell
	})

	t.Run("Synapse map is initialized and can add synapse immediately", func(t *testing.T) {
		synapse := NewSynapse(&original)
		original.Synapses[synapse.ID] = &synapse
	})

	t.Run("Calling Grow() with all empty values does not crash", func(t *testing.T) {
		original.Grow(0, 0, 0)
	})

}
