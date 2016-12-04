package potential

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NetworkGrow(t *testing.T) {
	t.Run("Grow() adds the right number of cells and synapses", func(t *testing.T) {
		n := NewNetwork()
		network := &n
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
		n := NewNetwork()
		network = &n

		// synapses are in (parens)
		//
		//
		//          (layer1A) - middle1 - (layer2C)
		//         /                               \
		// input -                                   - output
		//         \                               /
		//          (layer1B) - middle2 - (layer2D)

		input = NewCell()
		network.Cells[input.ID] = input
		output = NewCell()
		network.Cells[output.ID] = output

		middle1 = NewCell()
		network.Cells[middle1.ID] = middle1
		middle2 = NewCell()
		network.Cells[middle2.ID] = middle2

		// setup synapses
		layer1A = NewSynapse()
		network.Synapses[layer1A.ID] = layer1A
		layer1B = NewSynapse()
		network.Synapses[layer1B.ID] = layer1B
		layer2C = NewSynapse()
		network.Synapses[layer2C.ID] = layer2C
		layer2D = NewSynapse()
		network.Synapses[layer2D.ID] = layer2D

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
	t.Run("adds synapses when the number of connections is below the minimum", func(t *testing.T) {
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
		n := NewNetwork()
		network = &n

		var lastCell *Cell
		var lastHoppedCell *Cell

		input := NewCell()
		network.Cells[input.ID] = input
		input.Tag = "input"
		output := NewCell()
		output.Tag = "output"
		network.Cells[output.ID] = output

		lastCell = input

		// make a single path from input to output that is longer than the max hops
		// minimum value of 20
		for i := 0; i < 25; i++ {
			s := NewSynapse()
			network.Synapses[s.ID] = s
			c := NewCell()
			network.Cells[c.ID] = c

			s.FromNeuronAxon = lastCell.ID
			lastCell.AxonSynapses[s.ID] = true
			s.ToNeuronDendrite = c.ID
			c.DendriteSynapses[s.ID] = true
			c.Tag = strconv.Itoa(i)
			lastCell = c
			if i == 18 {
				lastHoppedCell = c
			}
		}

		// pre test checks
		// these next two cells will be linked together
		assert.Equal(t, 0, len(output.DendriteSynapses))
		assert.Equal(t, 1, len(lastHoppedCell.AxonSynapses))
		lastHoppedCell.Tag += " last hopped cell"

		endSynapses, addedSynapses := network.GrowPathBetween(input.ID, output.ID, 10)

		// test assertions
		assert.Equal(t, 0, len(endSynapses))
		assert.Equal(t, 10, len(addedSynapses))
		assert.Equal(t, 35, len(network.Synapses))
		assert.Equal(t, 10, len(output.DendriteSynapses))
		assert.Equal(t, 11, len(lastHoppedCell.AxonSynapses))
	})
	t.Run("does not panic on a large network with many synapses", func(t *testing.T) {
		n := NewNetwork()
		network = &n
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
		n := NewNetwork()
		network := &n
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
}

func Test_GrowPathBetween_Integrity(t *testing.T) {
	t.Run("passes the integrity after growing many deep paths on a large network", func(t *testing.T) {
		n := NewNetwork()
		network := &n
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
}
