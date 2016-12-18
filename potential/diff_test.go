package potential

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewDiff(t *testing.T) {
	t.Run("initializes maps and arrays so immediate assignment does not panic", func(t *testing.T) {
		n := NewNetwork()
		network := &n
		diff := NewDiff()

		cell := NewCell(network)
		synapse := NewSynapse(network)

		// tests are here
		diff.addedCells = append(diff.addedCells, cell)
		diff.addedSynapses = append(diff.addedSynapses, synapse)
	})
}

func Test_CloneNetwork(t *testing.T) {
	t.Run("cloning a network yields a valid copy", func(t *testing.T) {
		o := NewNetwork()
		original := &o

		beforeCell := NewCell(original)
		original.Cells[beforeCell.ID] = beforeCell
		beforeCell.Network = original

		cloned := CloneNetwork(original)
		// change something
		three := SynapseID(333)
		cloned.nextSynapsesToActivate[three] = true

		if cloned.nextSynapsesToActivate[three] == original.nextSynapsesToActivate[three] {
			t.Error("changing props of cloned network should not change original")
		}

		if &cloned.Cells[beforeCell.ID].Network == &original.Cells[beforeCell.ID].Network {
			t.Error("cloned and original cells should not point to the same network")
		}

		assert.ObjectsAreEqualValues(beforeCell, cloned.Cells[beforeCell.ID])
		c1 := original.Cells[beforeCell.ID]
		c2 := cloned.Cells[beforeCell.ID]

		if &c1 == &c2 {
			t.Error("pointers to same ID cell between cloned networks should not be equal")
		}
	})
	t.Run("all cells and synapses have valid connections on the cloned network", func(t *testing.T) {
		o := NewNetwork()
		original := &o
		o.Grow(1000, 10, 1000)

		ok, report := CheckIntegrity(original)
		assert.True(t, ok, "A freshly grown network had bad integrity", report)

		cloned := CloneNetwork(original)

		ok, report = CheckIntegrity(cloned)
		assert.True(t, ok, "A freshly cloned network had bad integrity", report)
	})
	t.Run("cloned network has the same synapses and cells", func(t *testing.T) {
		o := NewNetwork()
		original := &o
		o.Grow(100, 10, 100)

		cloned := CloneNetwork(original)

		assert.ObjectsAreEqualValues(*original, *cloned)
		assert.Equal(t, len(original.Synapses), len(cloned.Synapses))
	})
}

func Test_DiffNetworks(t *testing.T) {
	t.Run("synapse millivolt diff is NEW minus OLD", func(t *testing.T) {
		o := NewNetwork()
		original := &o
		syn1 := NewSynapse(original)
		syn1.Millivolts = 5
		// add cell for integrity
		c := NewCell(original)
		syn1.ToNeuronDendrite = c.ID
		syn1.FromNeuronAxon = c.ID
		c.AxonSynapses[syn1.ID] = true
		c.DendriteSynapses[syn1.ID] = true

		cloned := CloneNetwork(original)

		cloned.Synapses[syn1.ID].Millivolts = 10

		diff := DiffNetworks(original, cloned)
		assert.Equal(t, len(diff.synapseDiffs), 1, "Should be 1 synapse diff")
		assert.Equal(t, diff.synapseDiffs[syn1.ID], int8(5), "Synapse diff should be NEW - OLD")
	})
	t.Run("diffing unrelated networks works", func(t *testing.T) {
		n1 := NewNetwork()
		n2 := NewNetwork()
		net1 := &n1
		net2 := &n2

		n1.Grow(2000, 10, 1000)
		n2.Grow(2000, 10, 1000)

		diff := DiffNetworks(net1, net2)

		ApplyDiff(diff, net1)
		ok, report := CheckIntegrity(net1)
		assert.True(t, ok, "merging unrelated networks failed integrity check", report)

		// make sure a prune works afterward
		half := len(net1.Synapses) / 2
		iter := 0
		for _, synapse := range net1.Synapses {
			iter++
			if iter > half {
				break
			}
			synapse.ActivationHistory = 100
		}
		net1.Prune()
		ok, report = CheckIntegrity(net1)
		assert.True(t, ok, "pruning failed after merging totally different networks", report)
	})
}

func Test_ApplyDiff(t *testing.T) {
	t.Run("synapse millivolts are properly applied", func(t *testing.T) {
		o := NewNetwork()
		original := &o
		syn1 := NewSynapse(original)
		syn1.Millivolts = 7

		// must link to a cell for integrity check
		cell := NewCell(original)
		syn1.FromNeuronAxon = cell.ID
		syn1.ToNeuronDendrite = cell.ID

		cloned := CloneNetwork(original)

		cloned.Synapses[syn1.ID].Millivolts = 14
		ok, report := CheckIntegrity(cloned)
		if !ok {
			cloned.PrintCells()
		}
		assert.True(t, ok, "cloned network has bad integrity", report)

		diff := DiffNetworks(original, cloned)
		ApplyDiff(diff, original)

		assert.Equal(t, int8(14), original.Synapses[syn1.ID].Millivolts,
			"synapse millivolts failed to apply")
	})
	t.Run("all synapses and cells have valid connections after applying a diff", func(t *testing.T) {
		n1 := NewNetwork()
		net1 := &n1
		net1.Grow(50, 5, 50)
		net2 := CloneNetwork(net1)

		// make the networks different
		net2.GrowRandomNeurons(200, 10)
		net2.GrowRandomSynapses(100)

		// precheck
		ok, report := CheckIntegrity(net1)
		assert.True(t, ok, report)
		ok, report = CheckIntegrity(net2)
		assert.True(t, ok, report)

		// main thing
		diff := DiffNetworks(net1, net2)
		ApplyDiff(diff, net1)

		// assertions
		assert.Equal(t, len(net1.Synapses), (50*5)+50+(200*10)+100)
		ok, report = CheckIntegrity(net1)
		assert.True(t, ok, report)
	})
	t.Run("adds new cell to the network", func(t *testing.T) {
	})
}

func Test_copyCellToNetwork(t *testing.T) {
	t.Run("copying a cell sets the pointer to a new network", func(t *testing.T) {
		on := NewNetwork()
		originalNetwork := &on
		nn := NewNetwork()
		newNetwork := &nn
		originalCell := NewCell(originalNetwork)
		copyCellToNetwork(originalCell, newNetwork)
		copiedCell, exists := newNetwork.Cells[originalCell.ID]
		if !exists {
			t.Error("cell not copied to new network")
		}
		if copiedCell.Network != newNetwork {
			t.Error("cell copied but new network prop not set to new network pointer")
		}
	})
	t.Run("generates new cell ID and updates IDs when it already exists on the new network", func(t *testing.T) {})
	t.Run("a re-IDd+applied cell does not disrupt network integrity", func(t *testing.T) {})
}

func Test_copySynapseToNetwork(t *testing.T) {
	t.Run("generates new synapse ID and updates IDs when it already exists on the new network", func(t *testing.T) {})
	t.Run("a re-IDd+applied synapse does not disrupt network integrity", func(t *testing.T) {})
}
