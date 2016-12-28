package potential

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_PruneNetwork(t *testing.T) {
	t.Run("pruning a very large network maintains integrity", func(t *testing.T) {
		network := NewNetwork()
		network.Grow(4000, 10, 4000)
		ok, report := CheckIntegrity(network)
		assert.Equal(t, true, ok, report)

		// add firing to cells
		half := len(network.Synapses) / 2
		iter := 0
		for _, synapse := range network.Synapses {
			iter++
			if iter > half {
				break
			}
			synapse.ActivationHistory = 100
		}

		network.Prune()
		ok, report = CheckIntegrity(network)
		assert.Equal(t, true, ok, report)
		assert.Equal(t, 4000, len(network.Cells), "did not prune the right amount of cells")
		assert.Equal(t, 22000, len(network.Synapses), "did not prune the right number of synapses")

		t.Run("and can be diffed onto another network", func(t *testing.T) {
			otherNetwork := NewNetwork()
			otherNetwork.Grow(4000, 5, 1000)
			half = len(otherNetwork.Synapses) / 2
			iter = 0
			for _, synapse := range otherNetwork.Synapses {
				iter++
				if iter > half {
					break
				}
				synapse.ActivationHistory = 100
			}
			otherNetwork.Prune()
			ok, _ := CheckIntegrity(otherNetwork)

			// precheck
			assert.Equal(t, true, ok)

			diff := DiffNetworks(otherNetwork, network)
			ApplyDiff(diff, otherNetwork)

			ok, report := CheckIntegrity(otherNetwork)
			assert.Equal(t, true, ok)
			if !ok {
				report.Print()
			}

		})
	})

	t.Run("pruning degrades synapse millivolts toward zero when synapse activated less than desired times", func(t *testing.T) {
		network := NewNetwork()
		synapsePositive := NewSynapse(network)
		synapseNegative := NewSynapse(network)
		synapsePositive.Millivolts = 5
		synapseNegative.Millivolts = -5
		cell := NewCell(network)
		cell.AxonSynapses[synapsePositive.ID] = true
		cell.AxonSynapses[synapseNegative.ID] = true
		cell.DendriteSynapses[synapsePositive.ID] = true
		cell.DendriteSynapses[synapseNegative.ID] = true
		synapsePositive.FromNeuronAxon = cell.ID
		synapsePositive.ToNeuronDendrite = cell.ID
		synapseNegative.FromNeuronAxon = cell.ID
		synapseNegative.ToNeuronDendrite = cell.ID

		synapsePositive.ActivationHistory = defaultSynapseMinFireThreshold - 1
		synapseNegative.ActivationHistory = defaultSynapseMinFireThreshold - 1

		network.Prune()

		assert.Equal(t, int8(2), synapsePositive.Millivolts)
		assert.Equal(t, int8(-2), synapseNegative.Millivolts)
	})

	t.Run("pruning sets synapse millivolts to zero when synapse activated less than desired times and is half distance to zero is 2 or less", func(t *testing.T) {
		network := NewNetwork()
		synapsePositive := NewSynapse(network)
		synapseNegative := NewSynapse(network)
		synapsePositive.Millivolts = 3
		synapseNegative.Millivolts = -3
		cell := NewCell(network)
		cell.AxonSynapses[synapsePositive.ID] = true
		cell.AxonSynapses[synapseNegative.ID] = true
		cell.DendriteSynapses[synapsePositive.ID] = true
		cell.DendriteSynapses[synapseNegative.ID] = true
		synapsePositive.FromNeuronAxon = cell.ID
		synapsePositive.ToNeuronDendrite = cell.ID
		synapseNegative.FromNeuronAxon = cell.ID
		synapseNegative.ToNeuronDendrite = cell.ID

		synapsePositive.ActivationHistory = defaultSynapseMinFireThreshold - 1
		synapseNegative.ActivationHistory = defaultSynapseMinFireThreshold - 1

		network.Prune()

		assert.Equal(t, int8(0), synapsePositive.Millivolts)
		assert.Equal(t, int8(0), synapseNegative.Millivolts)
	})
	t.Run("pruning does not bump millivolts outside int8 bounds", func(t *testing.T) {
		network := NewNetwork()
		synapsePositive := NewSynapse(network)
		synapseNegative := NewSynapse(network)
		synapsePositive.Millivolts = 126
		synapseNegative.Millivolts = -127
		cell := NewCell(network)
		cell.AxonSynapses[synapsePositive.ID] = true
		cell.AxonSynapses[synapseNegative.ID] = true
		cell.DendriteSynapses[synapsePositive.ID] = true
		cell.DendriteSynapses[synapseNegative.ID] = true
		synapsePositive.FromNeuronAxon = cell.ID
		synapsePositive.ToNeuronDendrite = cell.ID
		synapseNegative.FromNeuronAxon = cell.ID
		synapseNegative.ToNeuronDendrite = cell.ID

		synapsePositive.ActivationHistory = defaultSynapseMinFireThreshold + 5
		synapseNegative.ActivationHistory = defaultSynapseMinFireThreshold + 5

		network.Prune()

		assert.Equal(t, int8(125), synapsePositive.Millivolts)
		assert.Equal(t, int8(-126), synapseNegative.Millivolts)
	})
	t.Run("pruning removes synapses that did not activate", func(t *testing.T) {
		// setup
		network := NewNetwork()
		s1 := NewSynapse(network)
		s2 := NewSynapse(network)
		c3 := NewCell(network)
		c4 := NewCell(network)

		s1.FromNeuronAxon = c3.ID
		c3.AxonSynapses[s1.ID] = true

		s2.ToNeuronDendrite = c4.ID
		c4.DendriteSynapses[s2.ID] = true

		s2.FromNeuronAxon = c3.ID
		c3.AxonSynapses[s2.ID] = true

		s1.ToNeuronDendrite = c4.ID
		c4.DendriteSynapses[s1.ID] = true

		// prechecks
		integrity, _ := CheckIntegrity(network)
		assert.Equal(t, true, integrity)
		assert.Equal(t, 2, len(network.Cells))
		assert.Equal(t, 2, len(network.Synapses))

		// test
		network.Prune()
		assert.Equal(t, 0, len(network.Synapses))
	})

	t.Run("pruning bumps sufficiently activated synapse millivolts by the learn rate", func(t *testing.T) {

	})
}

func Test_PruneSynapse(t *testing.T) {
	var network *Network
	var synapse *Synapse
	var cell1, cell2 *Cell
	before := func() {
		// setup
		network = NewNetwork()
		synapse = NewSynapse(network)
		network.Synapses[synapse.ID] = synapse
		synapse.Network = network
		// cell 1 fires into cell 2
		cell1 = NewCell(network)
		cell2 = NewCell(network)
		cell1.AxonSynapses[synapse.ID] = true
		cell2.DendriteSynapses[synapse.ID] = true
		synapse.FromNeuronAxon = cell1.ID
		synapse.ToNeuronDendrite = cell2.ID
	}

	t.Run("removes synapse from the network", func(t *testing.T) {
		before()
		network.PruneSynapse(synapse.ID)
		_, ok := network.Synapses[synapse.ID]
		assert.Equal(t, false, ok, "synapse not removed from network during PruneNetwork")
	})
	t.Run("maintains integrity after removal", func(t *testing.T) {
		before()
		network.PruneSynapse(synapse.ID)
		ok, _ := CheckIntegrity(network)
		assert.Equal(t, true, ok)
	})
	t.Run("removes synapses from the actual network cells (not copying)", func(t *testing.T) {
		before()
		network.PruneSynapse(synapse.ID)
		_, ok := cell1.AxonSynapses[synapse.ID]
		if ok {
			panic("synapse not removed from axon side when pruned")
		}
		_, ok = cell1.DendriteSynapses[synapse.ID]
		if ok {
			panic("synapse not removed from dendrite side when pruned")
		}
	})
	t.Run("when cells have no synapses, it removes them too", func(t *testing.T) {
		before()
		network.PruneSynapse(synapse.ID)
		_, ok := network.Cells[cell1.ID]
		assert.Equal(t, 0, len(network.Cells), "network has too many cells after pruning synapse")

		assert.Equal(t, false, ok, "cell1 not removed from network when synapses were zero during synapse prune")

		_, ok = network.Cells[cell2.ID]
		assert.Equal(t, false, ok, "cell2 not removed from network when synapses were zero during synapse prune")
	})
	t.Run("does not crash when synapse does not exist", func(t *testing.T) {
		n := NewNetwork()
		n.PruneSynapse(54)
	})
	t.Run("removeSynapseFromCell panics when cell does not exist", func(t *testing.T) {
		n := NewNetwork()
		s := NewSynapse(n)
		c := NewCell(network)
		s.ToNeuronDendrite = 732
		s.FromNeuronAxon = c.ID
		// cell does not reference back
		assert.Panics(t, func() {
			network.removeSynapseFromCell(s.ID, 732, false)
		})
	})
}

func Test_PruneCell(t *testing.T) {
	t.Run("does not crash when cell is not on network", func(t *testing.T) {
		network := NewNetwork()
		NewCell(network)
		network.PruneCell(1)
	})
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
