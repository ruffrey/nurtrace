package potential

import (
	"strconv"
	"testing"

	"github.com/ruffrey/nurtrace/laws"

	"github.com/stretchr/testify/assert"
)

func _synapseInSlice(a SynapseID, list []SynapseID) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func Test_ConsolidateSynapses(t *testing.T) {
	t.Run("synapseSignature concats neuron IDs together", func(t *testing.T) {
		network := NewNetwork()
		s1 := NewSynapse(network)

		s1.FromNeuronAxon = CellID(3333)
		s1.ToNeuronDendrite = CellID(4444)
		assert.Equal(t, "3333-4444", synapseSignature(s1))
	})
	t.Run("findDupeSynapses returns groupings of synapses with duplicate neuron connections", func(t *testing.T) {
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
		actual := result[expectedSig]

		assert.Equal(t, 2, len(actual))
		assert.Equal(t, true, _synapseInSlice(s1.ID, actual))
		assert.Equal(t, true, _synapseInSlice(s2.ID, actual))
	})
	t.Run("dedupeSynapses removes duplicate synapses from a network and sums their Millivolts", func(t *testing.T) {
		var network *Network
		var cellA *Cell
		var cellB *Cell
		var cellC *Cell
		var cellD *Cell
		var s1dupe *Synapse
		var s2dupe *Synapse

		beforeEach := func() {
			network = NewNetwork()
			cellA = NewCell(network)
			cellB = NewCell(network)
			cellC = NewCell(network)
			cellD = NewCell(network)

			// two dupe synapses
			s1dupe = network.linkCells(cellA.ID, cellB.ID)
			s2dupe = network.linkCells(cellA.ID, cellB.ID)

			// four non-dupe synapses
			network.linkCells(cellA.ID, cellC.ID)
			network.linkCells(cellB.ID, cellD.ID)
			network.linkCells(cellA.ID, cellD.ID)
			network.linkCells(cellD.ID, cellB.ID)

			assert.Equal(t, 6, len(network.Synapses)) // sanity
		}
		t.Run("duplicate synapses positive", func(t *testing.T) {
			beforeEach()

			s1dupe.Millivolts = 25
			s2dupe.Millivolts = 41

			dupes := dupeSynapses{s1dupe.ID, s2dupe.ID}
			dedupeSynapses(dupes, network)

			assert.Equal(t, 5, len(network.Synapses))
			assert.Equal(t, int16(66), s1dupe.Millivolts)

			stillHasS2 := network.synExists(s2dupe.ID)
			assert.Equal(t, false, stillHasS2,
				"Did not remove synapse from network!")
		})
		t.Run("duplicate synapses negative", func(t *testing.T) {
			beforeEach()

			s1dupe.Millivolts = -35
			s2dupe.Millivolts = -2

			dupes := dupeSynapses{s1dupe.ID, s2dupe.ID}
			dedupeSynapses(dupes, network)

			assert.Equal(t, 5, len(network.Synapses))
			assert.Equal(t, int16(-37), s1dupe.Millivolts)
		})
		t.Run("duplicate synapses mixed signs", func(t *testing.T) {
			beforeEach()

			s1dupe.Millivolts = -12
			s2dupe.Millivolts = 19

			dupes := dupeSynapses{s1dupe.ID, s2dupe.ID}
			dedupeSynapses(dupes, network)

			assert.Equal(t, 5, len(network.Synapses))
			assert.Equal(t, int16(7), s1dupe.Millivolts)
		})
		t.Run("many dupes with mixed signs", func(t *testing.T) {
			beforeEach()

			s3dupe := network.linkCells(cellA.ID, cellB.ID)
			s4dupe := network.linkCells(cellA.ID, cellB.ID)
			s5dupe := network.linkCells(cellA.ID, cellB.ID)
			s6dupe := network.linkCells(cellA.ID, cellB.ID)

			s1dupe.Millivolts = 3
			s2dupe.Millivolts = -5
			s3dupe.Millivolts = 19
			s4dupe.Millivolts = -12
			s5dupe.Millivolts = 39
			s6dupe.Millivolts = 105

			assert.Equal(t, 10, len(network.Synapses))

			dupes := findDupeSynapses(network)
			sig := synapseSignature(s1dupe)
			assert.Equal(t, 6, len(dupes[sig]))
			leftovers := dedupeSynapses(dupes[sig], network)
			assert.Equal(t, 1, len(leftovers))

			assert.Equal(t, 5, len(network.Synapses))
			assert.Equal(t, int16(149), network.getSyn(leftovers[0]).Millivolts)
		})
		t.Run("dupes overflowing +int16 keeps more synapses", func(t *testing.T) {
			beforeEach()

			s3dupe := network.linkCells(cellA.ID, cellB.ID)

			s1dupe.Millivolts = 20000
			s2dupe.Millivolts = 30000
			s3dupe.Millivolts = 61
			assert.Equal(t, 7, len(network.Synapses))

			leftovers := dedupeSynapses(dupeSynapses{s1dupe.ID, s2dupe.ID, s3dupe.ID}, network)
			assert.Equal(t, 2, len(leftovers))

			assert.Equal(t, 6, len(network.Synapses))
			assert.Equal(t, laws.ActualSynapseMax, network.getSyn(leftovers[0]).Millivolts)
			assert.Equal(t, int16(17297), network.getSyn(leftovers[1]).Millivolts)
		})
		t.Run("dupes overflowing -int16 keeps more synapses", func(t *testing.T) {
			beforeEach()

			s3dupe := network.linkCells(cellA.ID, cellB.ID)

			s1dupe.Millivolts = -20000
			s2dupe.Millivolts = -30000
			s3dupe.Millivolts = -61
			assert.Equal(t, 7, len(network.Synapses))

			leftovers := dedupeSynapses(dupeSynapses{s1dupe.ID, s2dupe.ID, s3dupe.ID}, network)
			assert.Equal(t, 2, len(leftovers))

			assert.Equal(t, 6, len(network.Synapses))
			assert.Equal(t, laws.ActualSynapseMin, network.getSyn(leftovers[0]).Millivolts)
			assert.Equal(t, int16(-17296), network.getSyn(leftovers[1]).Millivolts)
		})
	})
}
