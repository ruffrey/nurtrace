package potential

import "testing"

func Test_NewNetwork(t *testing.T) {
	original := NewNetwork()

	t.Run("Cell map is initialized and can add cell immediately", func(t *testing.T) {
		cell := NewCell(&original)
		original.Cells[cell.ID] = &cell
	})

	t.Run("Synapse map is initialized and can add synapse immediately", func(t *testing.T) {
		synapse := NewSynapse(&original)
		original.Synapses[synapse.ID] = &synapse
	})

	t.Run("Calling Grow() with all empty values does not crash", func(t *testing.T) {
		original.Grow(0, 0, 0)
	})

}

func Test_PruneSynapse(t *testing.T) {
	var network Network
	var synapse Synapse
	var cell1, cell2 Cell
	before := func() {
		// setup
		network = NewNetwork()
		synapse = NewSynapse(&network)
		// cell 1 fires into cell 2
		cell1 = NewCell(&network)
		cell2 = NewCell(&network)
		cell1.AxonSynapses[synapse.ID] = true
		cell2.DendriteSynapses[synapse.ID] = true
		synapse.FromNeuronAxon = cell1.ID
		synapse.ToNeuronDendrite = cell2.ID
	}

	t.Run("removes synapse from the network", func(t *testing.T) {
		before()
		network.PruneSynapse(&synapse)
		_, ok := network.Synapses[synapse.ID]
		if ok {
			panic("synapse not removed from network during PruneNetwork")
		}
	})
	t.Run("removes synapses from the actual network cells (not copying)", func(t *testing.T) {
		before()
		network.PruneSynapse(&synapse)
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
		network.PruneSynapse(&synapse)
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
