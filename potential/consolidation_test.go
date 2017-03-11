package potential

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ConsolidateSynapses(t *testing.T) {
	t.Run("synapseSignature", func(t *testing.T) {
		network := NewNetwork()
		s1 := NewSynapse(network)

		s1.FromNeuronAxon = CellID(3333)
		s1.ToNeuronDendrite = CellID(4444)
		assert.Equal(t, "3333-4444", synapseSignature(s1))
	})
	t.Run("findDupeSynapses", func(t *testing.T) {
		network := NewNetwork()
		cellA := NewCell(network)
		cellB := NewCell(network)
		cellC := NewCell(network)
		cellD := NewCell(network)

		s1 := network.linkCells(cellA.ID, cellB.ID)
		s2 := network.linkCells(cellA.ID, cellB.ID)
		assert.Equal(t, 2, len(network.Synapses)) // sanity

		network.linkCells(cellA.ID, cellC.ID)
		network.linkCells(cellB.ID, cellD.ID)
		assert.Equal(t, 4, len(network.Synapses)) // sanity

		result := findDupeSynapses(network)

		assert.Equal(t, 1, len(result))

		expectedSig := strconv.Itoa(int(cellA.ID)) + "-" + strconv.Itoa(int(cellB.ID))
		expected := []SynapseID{s1.ID, s2.ID}
		assert.EqualValues(t, expected, result[expectedSig])
	})
	t.Run("dedupeSynapses", func(t *testing.T) {

	})
}
