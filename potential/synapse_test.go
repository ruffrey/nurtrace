package potential

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

/*func Test_SynapseActivateNotExist(t *testing.T) {
	var network *Network
	before := func() {
		network = NewNetwork()
	}

	t.Run("when synapse references the network but its dendrite not in cells it returns error", func(t *testing.T) {
		before()

		synapse := NewSynapse(network)
		synapse.ToNeuronDendrite = NewCellID()

		_, err := synapse.Activate()
		assert.Error(t, err, "err should exist when activating bad dendrite")
	})

	t.Run("when syanpse and cell exist, activating a synapse with low millivolts does not fire the cell", func(t *testing.T) {
		before()

		synapse := NewSynapse(network)
		cell := NewCell(network)

		synapse.ToNeuronDendrite = cell.ID
		synapse.Millivolts = 1

		// base state
		assert.Equal(t, false, cell.WasFired, "brand new cell should have WasFired==false")
		assert.Equal(t, cellRestingVoltage, cell.Voltage, "brand new cell should have resting voltage")

		didFire, err := synapse.Activate()
		assert.Nil(t, err, "no error when activating dendrite")

		assert.Equal(t, false, didFire, "cell should not have fired with such low voltage")
		assert.Equal(t, false, cell.WasFired,
			"new cell should be fired after its dendrite synapse activates")

		assert.Equal(t, int8(cellRestingVoltage+1), cell.Voltage, "activated cell should have voltage applied")
	})

	t.Run("when syanpse and cell exist, activating a synapse with high millivolts fires the cell", func(t *testing.T) {
		before()

		synapse := NewSynapse(network)

		cell := NewCell(network)

		synapse.ToNeuronDendrite = cell.ID
		synapse.Millivolts = 100

		// base state
		assert.Equal(t, false, cell.WasFired, "brand new cell should have WasFired==false")

		didFire, err := synapse.Activate()
		assert.Nil(t, err, "no error when activating dendrite")
		assert.Equal(t, true, didFire, "cell should have fired due to high voltage")

		assert.Equal(t, true, cell.WasFired,
			"new cell should be fired after its dendrite synapse activates")
	})
}*/

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
		s1.Millivolts = 125
		cell.addAxon(s1.ID)
		cell.addDendrite(s1.ID)
		s1.reinforce()
		assert.Equal(t, int8(62), s1.Millivolts)
		assert.Equal(t, 2, len(network.Synapses))

		s1.Millivolts = 126
		s1.reinforce()
		assert.Equal(t, int8(62), s1.Millivolts)
		assert.Equal(t, 3, len(network.Synapses))

		s2 := NewSynapse(network)
		s2.Millivolts = -126
		cell.addAxon(s2.ID)
		cell.addDendrite(s2.ID)
		s2.reinforce()
		assert.Equal(t, int8(-62), s2.Millivolts)
		assert.Equal(t, 5, len(network.Synapses))
		s2.Millivolts = -127
		s2.reinforce()
		assert.Equal(t, int8(-62), s2.Millivolts)
		assert.Equal(t, 6, len(network.Synapses))
	})
}
