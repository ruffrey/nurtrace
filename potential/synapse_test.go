package potential

import (
	"testing"

	"github.com/ruffrey/nurtrace/laws"

	"github.com/stretchr/testify/assert"
)

func Test_SynapseStringer(t *testing.T) {
	t.Run("String works without crashing", func(t *testing.T) {
		network := NewNetwork()
		synapse := NewSynapse(network)
		s := synapse.String()
		assert.NotEmpty(t, s)
	})
}

func Test_SynapseActivate(t *testing.T) {

}

func Test_SynapseReinforce(t *testing.T) {
	t.Run("reinforcing a negative synapse will make it more negative", func(t *testing.T) {
		network := NewNetwork()
		synapse := NewSynapse(network)
		synapse.Millivolts = -4
		synapse.reinforce()

		assert.Equal(t, -4-laws.SynapseLearnRate, synapse.Millivolts)
	})
	t.Run("reinforcing a positive synapse will make it more positive", func(t *testing.T) {
		network := NewNetwork()
		synapse := NewSynapse(network)
		synapse.Millivolts = 3
		synapse.reinforce()

		assert.Equal(t, 3+laws.SynapseLearnRate, synapse.Millivolts)
	})
	t.Run("reinforcing a synapse to its limit will not overflow integer and will add new duplicate synapses", func(t *testing.T) {
		network := NewNetwork()

		cell := NewCell(network)

		s1 := NewSynapse(network)
		s1.Millivolts = 32765
		cell.addAxon(s1.ID)
		cell.addDendrite(s1.ID)
		s1.reinforce()
		assert.Equal(t, int16(16382), s1.Millivolts)
		assert.Equal(t, 2, len(network.Synapses))

		s1.Millivolts = 32765
		s1.reinforce()
		assert.Equal(t, int16(16382), s1.Millivolts)
		assert.Equal(t, 3, len(network.Synapses))

		s2 := NewSynapse(network)
		s2.Millivolts = -32765
		cell.addAxon(s2.ID)
		cell.addDendrite(s2.ID)
		s2.reinforce()
		assert.Equal(t, int16(-16382), s2.Millivolts)
		assert.Equal(t, 5, len(network.Synapses))
		s2.Millivolts = -32765
		s2.reinforce()
		assert.Equal(t, int16(-16382), s2.Millivolts)
		assert.Equal(t, 6, len(network.Synapses))
	})
}

func Test_PruneSynapse(t *testing.T) {
	var network *Network
	var synapse *Synapse
	var cell1, cell2 *Cell
	before := func() {
		// setup
		network = NewNetwork()
		// cell 1 fires into cell 2
		cell1 = NewCell(network)
		cell2 = NewCell(network)
		synapse = network.linkCells(cell1.ID, cell2.ID)
	}

	t.Run("removes synapse from the network", func(t *testing.T) {
		before()
		network.PruneSynapse(synapse.ID)
		ok := network.synExists(synapse.ID)
		assert.Equal(t, false, ok, "synapse not removed from network during PruneNetwork")
	})
	t.Run("maintains integrity after removal", func(t *testing.T) {
		before()
		network.PruneSynapse(synapse.ID)
		ok, integrity := CheckIntegrity(network)
		assert.Equal(t, true, ok, integrity)
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
		ok := network.cellExists(cell1.ID)

		assert.Equal(t, false, ok, "cell1 not removed from network when synapses were zero during synapse prune")

		ok = network.cellExists(cell2.ID)
		assert.Equal(t, false, ok, "cell2 not removed from network when synapses were zero during synapse prune")
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
