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
		cell1 = NewCell()
		network.Cells[cell1.ID] = cell1
		cell1.Network = network
		cell2 = NewCell()
		network.Cells[cell2.ID] = cell2
		cell2.Network = network
		cell1.AxonSynapses[synapse.ID] = true
		cell2.DendriteSynapses[synapse.ID] = true
		synapse.FromNeuronAxon = cell1.ID
		synapse.ToNeuronDendrite = cell2.ID
	}

	t.Run("removes synapse from the network", func(t *testing.T) {
		before()
		network.PruneSynapse(synapse.ID)
		_, ok := network.Synapses[synapse.ID]
		if ok {
			panic("synapse not removed from network during PruneNetwork")
		}
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
		if ok {
			panic("cell1 not removed from network when synapses were zero during synapse prune")
		}
		_, ok = network.Cells[cell2.ID]
		if ok {
			panic("cell2 not removed from network when synapses were zero during synapse prune")
		}
	})
}
