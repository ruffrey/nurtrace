package potential

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_PruneSynapse(t *testing.T) {
	var network *Network
	var synapse *Synapse
	var cell1, cell2 *Cell
	before := func() {
		// setup
		n := NewNetwork()
		network = &n
		synapse = NewSynapse()
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
		ok, _ := checkIntegrity(network)
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
