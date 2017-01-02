package potential

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_backwardTraceFiringsGood(t *testing.T) {
	t.Run("returns a list of excitatory synapses between cells that fired", func(t *testing.T) {
		network := NewNetwork()
		c1 := NewCell(network)
		c2 := NewCell(network)
		c3 := NewCell(network)
		c4 := NewCell(network)
		c5 := NewCell(network)

		c1.WasFired = true
		c2.WasFired = true
		c3.WasFired = true
		c4.WasFired = true
		c5.WasFired = true

		s1 := NewSynapse(network)
		s2 := NewSynapse(network)
		s3 := NewSynapse(network)
		s4 := NewSynapse(network)

		// Network structure:
		//	   -> +s1 -> c2 -> +s2 ->
		//	c1			  c5
		//	   -> +s3 -> c4 -> -s4 ->

		s1.Millivolts = 50
		s2.Millivolts = 70
		s3.Millivolts = 50
		s4.Millivolts = -50

		// top path
		c1.addAxon(s1.ID)
		c2.addDendrite(s1.ID)
		c2.addAxon(s2.ID)
		c5.addDendrite(s2.ID)

		// bottom path
		c1.addAxon(s3.ID)
		c4.addDendrite(s3.ID)
		c4.addAxon(s4.ID)
		c5.addDendrite(s4.ID)

		goodSynapses := backwardTraceFirings(network, c5.ID, c1.ID)
		assert.Equal(t, 2, len(goodSynapses))
		_, exists := goodSynapses[s2.ID]
		assert.True(t, exists)
		_, exists = goodSynapses[s1.ID]
		assert.True(t, exists)
	})
}

func Test_backwardTraceFiringsBad(t *testing.T) {
	t.Run("returns a list of excitatory synapses that fired the wrong cell", func(t *testing.T) {
		network := NewNetwork()
		c1 := NewCell(network)
		c2 := NewCell(network)
		c3 := NewCell(network)
		c4 := NewCell(network)
		c5 := NewCell(network)
		c6 := NewCell(network)

		c1.WasFired = true
		c2.WasFired = true
		c3.WasFired = true
		c4.WasFired = true
		c5.WasFired = true
		c6.WasFired = true

		s1 := NewSynapse(network)
		s2 := NewSynapse(network)
		s3 := NewSynapse(network)
		s4 := NewSynapse(network)

		// Network structure:
		//	   -> +s1 -> c2 -> +s2 -> c5
		//	c1
		//	   -> +s3 -> c4 -> +s4 -> c6

		s1.Millivolts = 50
		s2.Millivolts = 70
		s3.Millivolts = 50
		s4.Millivolts = 50

		// top path
		c1.addAxon(s1.ID)
		c2.addDendrite(s1.ID)
		c2.addAxon(s2.ID)
		c5.addDendrite(s2.ID)

		// bottom path
		c1.addAxon(s3.ID)
		c4.addDendrite(s3.ID)
		c4.addAxon(s4.ID)
		c6.addDendrite(s4.ID)

		inputCells := make(map[CellID]bool)
		inputCells[c1.ID] = true

		unexpectedOutputCells := make(map[CellID]bool)
		unexpectedOutputCells[c6.ID] = true
		goodSynapses := backwardTraceFirings(network, c5.ID, c1.ID)
		badSynapses := backwardTraceNoise(network, inputCells, unexpectedOutputCells, goodSynapses)

		assert.Equal(t, 2, len(goodSynapses), "should be 2 goodSynapses")
		assert.Equal(t, 2, len(badSynapses), "should be 2 badSynapses")
		_, exists := badSynapses[s4.ID]
		assert.True(t, exists)
		_, exists = badSynapses[s3.ID]
		assert.True(t, exists)

	})
}

func Test_applyBacktrace(t *testing.T) {
	t.Run("it increases millivolts away from zero on good path", func(t *testing.T) {

	})
}
