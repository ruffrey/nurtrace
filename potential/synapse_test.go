package potential

import (
	"testing"

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

		assert.Equal(t, -4-synapseLearnRate, synapse.Millivolts)
	})
	t.Run("reinforcing a positive synapse will make it more positive", func(t *testing.T) {
		network := NewNetwork()
		synapse := NewSynapse(network)
		synapse.Millivolts = 3
		synapse.reinforce()

		assert.Equal(t, 3+synapseLearnRate, synapse.Millivolts)
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
