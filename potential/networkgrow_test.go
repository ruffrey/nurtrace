package potential

import "testing"

func Test_GrowPathBetween(t *testing.T) {
	var network *Network

	before := func() {
		n := NewNetwork()
		network = &n
	}

	t.Run("finds all synapses connected to the end cell", func(t *testing.T) {
		before()

	})
	t.Run("does not find synapses past the maxHops", func(t *testing.T) {
		before()

	})
	t.Run("does not panic on a large network with many synapses", func(t *testing.T) {
		before()
		network.Grow(5000, 10, 0)
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
	t.Run("adds synapses when the number of connections is below the minimum", func(t *testing.T) {
		before()

	})
	t.Run("does not add synapses when the number of connections is at the minimum", func(t *testing.T) {
		before()

	})
}
