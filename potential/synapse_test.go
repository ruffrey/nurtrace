package potential

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SynapseActivateNotExist(t *testing.T) {
	var network *Network
	before := func() {
		n := NewNetwork()
		network = &n
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
		assert.Equal(t, int8(-70), cell.Voltage, "brand new cell should have voltage==-70")

		didFire, err := synapse.Activate()
		assert.Nil(t, err, "no error when activating dendrite")

		assert.Equal(t, false, didFire, "cell should not have fired with such low voltage")
		assert.Equal(t, false, cell.WasFired,
			"new cell should be fired after its dendrite synapse activates")

		assert.Equal(t, int8(-69), cell.Voltage, "activated cell should have voltage applied")
	})

	t.Run("when syanpse and cell exist, activating a synapse with high millivolts fires the cell", func(t *testing.T) {
		before()

		synapse := NewSynapse(network)

		cell := NewCell(network)

		synapse.ToNeuronDendrite = cell.ID
		synapse.Millivolts = 50

		// base state
		assert.Equal(t, false, cell.WasFired, "brand new cell should have WasFired==false")

		didFire, err := synapse.Activate()
		assert.Nil(t, err, "no error when activating dendrite")
		assert.Equal(t, true, didFire, "cell should have fired due to high voltage")

		assert.Equal(t, true, cell.WasFired,
			"new cell should be fired after its dendrite synapse activates")
	})
}

func Test_SynapseActivate(t *testing.T) {

}
