package potential

import "testing"

func Test_NetworkStep(t *testing.T) {
	t.Run("does not crash when synapses or cells not exist", func(t *testing.T) {
		network := NewNetwork()
		network.nextSynapsesToActivate[4] = true
		network.resetCellsOnNextStep[71] = true
		network.Step()
	})
	t.Run("does not crash when adding nonexistent synapse", func(t *testing.T) {
		network := NewNetwork()
		network.AddSynapseToNextStep(97)
		network.Step()
	})
	t.Skip("when firing one round results in firing the next round it returns true")
	t.Skip("when firing one round results in no new firings it returns false")
}
