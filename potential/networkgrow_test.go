package potential

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NetworkGrow(t *testing.T) {
	t.Run("Grow() adds the right number of cells and synapses", func(t *testing.T) {
		network := NewNetwork()
		network.Grow(50, 5, 200)

		assert.Equal(t, 50, len(network.Cells))
		assert.Equal(t, (50*5)+200, len(network.Synapses))
	})
}

func Test_GrowPathBetween(t *testing.T) {
	var network *Network
	var input *Cell
	var output *Cell
	var middle1 *Cell
	var middle2 *Cell
	var layer1A *Synapse
	var layer1B *Synapse
	var layer2C *Synapse
	var layer2D *Synapse

	before := func() {
		network = NewNetwork()

		// synapses are in (parens) and all positive mv
		//
		//
		//          (+layer1A) - middle1 - (+layer2C)
		//         /                               \
		// input -                                   - output
		//         \                               /
		//          (+layer1B) - middle2 - (+layer2D)

		input = NewCell(network)
		input.Tag = "input"
		output = NewCell(network)
		output.Tag = "output"

		middle1 = NewCell(network)
		middle1.Tag = "middle1"
		middle2 = NewCell(network)
		middle2.Tag = "middle2"

		// setup synapses
		layer1A = NewSynapse(network)
		layer1B = NewSynapse(network)
		layer2C = NewSynapse(network)
		layer2D = NewSynapse(network)

		// set synpase mv
		layer1A.Millivolts = 10
		layer1B.Millivolts = 10
		layer2C.Millivolts = 10
		layer2D.Millivolts = 10

		// linking synapses

		// top layer in diagram
		input.AxonSynapses[layer1A.ID] = true
		layer1A.FromNeuronAxon = input.ID
		middle1.DendriteSynapses[layer1A.ID] = true
		layer1A.ToNeuronDendrite = middle1.ID

		middle1.AxonSynapses[layer2C.ID] = true
		layer2C.FromNeuronAxon = middle1.ID
		output.DendriteSynapses[layer2C.ID] = true
		layer2C.ToNeuronDendrite = output.ID

		// bottom layer in diagram
		input.AxonSynapses[layer1B.ID] = true
		layer1B.FromNeuronAxon = input.ID
		middle2.DendriteSynapses[layer1B.ID] = true
		layer1B.ToNeuronDendrite = middle2.ID

		middle2.AxonSynapses[layer2D.ID] = true
		layer2D.FromNeuronAxon = middle2.ID
		output.DendriteSynapses[layer2D.ID] = true
		layer2D.ToNeuronDendrite = output.ID
	}

	t.Run("finds all synapses connected to the end cell and does not add extra synapses", func(t *testing.T) {
		before()
		endSynapses, addedSynapses := network.GrowPathBetween(input.ID, output.ID, 2)
		assert.Equal(t, 2, len(endSynapses))
		assert.Equal(t, true, endSynapses[layer2C.ID])
		assert.Equal(t, true, endSynapses[layer2D.ID])
		assert.Equal(t, 0, len(addedSynapses))
		assert.Equal(t, 4, len(network.Cells))
		assert.Equal(t, 4, len(network.Synapses))
	})
	t.Run("adds synapses and cells when the number of connections is below the minimum", func(t *testing.T) {
		before()

		endSynapses, addedSynapses := network.GrowPathBetween(input.ID, output.ID, 4)

		assert.Equal(t, 2, len(endSynapses))
		assert.Equal(t, true, endSynapses[layer2C.ID])
		assert.Equal(t, true, endSynapses[layer2D.ID])
		assert.Equal(t, 2, len(addedSynapses))
		assert.Equal(t, 4, len(network.Cells))
		assert.Equal(t, 6, len(network.Synapses))

	})
	t.Run("does not find synapses past the maxHops", func(t *testing.T) {
		before()
		network = NewNetwork()

		var lastCell *Cell

		input := NewCell(network)
		input.Tag = "input"
		output := NewCell(network)
		output.Tag = "output"

		lastCell = input

		// make a single path from input to output that is longer than the max hops
		// minimum value of 20
		for i := 0; i < 25; i++ {
			s := NewSynapse(network)
			c := NewCell(network)

			s.Millivolts = 100

			s.FromNeuronAxon = lastCell.ID
			lastCell.AxonSynapses[s.ID] = true
			s.ToNeuronDendrite = c.ID
			c.DendriteSynapses[s.ID] = true
			c.Tag = strconv.Itoa(i)
			lastCell = c
		}

		// pre test checks
		// these next two cells will be linked together
		assert.Equal(t, 0, len(output.DendriteSynapses))

		endSynapses, addedSynapses := network.GrowPathBetween(input.ID, output.ID, 10)

		// test assertions
		assert.Equal(t, 0, len(endSynapses))
		assert.Equal(t, 10, len(addedSynapses))
		assert.Equal(t, 35, len(network.Synapses))
		assert.Equal(t, 1, len(output.DendriteSynapses))
	})
	t.Run("does not panic on a large network with many synapses", func(t *testing.T) {
		network = NewNetwork()
		network.Grow(500, 10, 0)
		input := network.RandomCellKey()
		var output CellID
		for {
			output = network.RandomCellKey()
			if input != output {
				break
			}
		}
		network.GrowPathBetween(input, output, 10)
	})
	t.Run("changes the number of synapses during runs", func(t *testing.T) {
		network := NewNetwork()
		network.Grow(500, 0, 0)
		for i := 0; i < 10; i++ {
			input := network.RandomCellKey()
			var output CellID
			for {
				output = network.RandomCellKey()
				if input != output {
					break
				}
			}
			beforeCount := len(network.Synapses)
			network.GrowPathBetween(input, output, 10)
			assert.NotEqual(t, beforeCount, len(network.Synapses))
		}
	})
	t.Run("does not find negative synapses", func(t *testing.T) {

	})
}

func Test_GrowPathBetween_Integrity(t *testing.T) {
	t.Run("passes the integrity after growing many deep paths on a large network", func(t *testing.T) {
		network := NewNetwork()
		network.Grow(500, 5, 100)
		for i := 0; i < 10; i++ {
			input := network.RandomCellKey()
			var output CellID
			for {
				output = network.RandomCellKey()
				if input != output {
					break
				}
			}
			network.GrowPathBetween(input, output, 20)
		}
	})
	t.Run("does not add synapses when there are already enough", func(t *testing.T) {

	})
	t.Run("adds valid synapses and all cells have working connections after growth", func(t *testing.T) {

	})
}
