package potential

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_PruneNetwork(t *testing.T) {
	t.Run("pruning a network maintains integrity", func(t *testing.T) {
		n := NewNetwork()
		network := &n
		network.Grow(1000, 10, 1000)
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
		assert.Equal(t, 1000, len(network.Cells), "did not prune the right amount of cells")
		assert.Equal(t, 5500, len(network.Synapses), "did not prune the right number of synapses")

		t.Run("and can be diffed onto another network", func(t *testing.T) {
			on := NewNetwork()
			otherNetwork := &on
			otherNetwork.Grow(1000, 5, 100)
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
}

func Test_PruneSynapse(t *testing.T) {
	var network *Network
	var synapse *Synapse
	var cell1, cell2 *Cell
	before := func() {
		// setup
		n := NewNetwork()
		network = &n
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
}
