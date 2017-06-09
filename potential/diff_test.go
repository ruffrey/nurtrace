package potential

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewDiff(t *testing.T) {
	t.Run("initializes maps and arrays so immediate assignment does not panic", func(t *testing.T) {
		network := NewNetwork()
		diff := NewDiff()

		cell := NewCell(network)
		synapse := NewSynapse(network)

		// tests are here
		diff.addedCells[cell.ID] = cell
		diff.addedSynapses = append(diff.addedSynapses, synapse)
	})
}

func Test_CloneNetwork(t *testing.T) {
	t.Run("cloning a network yields a valid copy", func(t *testing.T) {
		original := NewNetwork()

		beforeCell := NewCell(original)
		afterCell := NewCell(original)
		original.linkCells(beforeCell.ID, afterCell.ID)

		cloned := CloneNetwork(original)
		// change something
		one := SynapseID(0)
		cloned.GetSyn(one).fireNextRound = true

		if cloned.GetSyn(one).fireNextRound == original.GetSyn(one).fireNextRound {
			t.Error("changing props of cloned network should not change original")
		}

		if &cloned.GetCell(beforeCell.ID).Network == &original.GetCell(beforeCell.ID).Network {
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
		original := NewNetwork()
		original.Grow(1000, 10, 1000)

		ok, report := CheckIntegrity(original)
		assert.True(t, ok, "A freshly grown network had bad integrity", report)

		cloned := CloneNetwork(original)

		ok, report = CheckIntegrity(cloned)
		assert.True(t, ok, "A freshly cloned network had bad integrity", report)
	})
	t.Run("cloned network has the same synapses and cells", func(t *testing.T) {
		original := NewNetwork()
		original.Grow(100, 10, 100)

		cloned := CloneNetwork(original)

		assert.ObjectsAreEqualValues(original.Cells, cloned.Cells)
		assert.ObjectsAreEqualValues(original.Synapses, cloned.Synapses)
		assert.Equal(t, len(original.Cells), len(cloned.Cells))
		assert.Equal(t, len(original.Synapses), len(cloned.Synapses))
	})
}

func Test_DiffNetworks(t *testing.T) {
	t.Run("synapse millivolt diff is NEW minus OLD", func(t *testing.T) {
		original := NewNetwork()
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
		assert.Equal(t, diff.synapseDiffs[syn1.ID], int16(5), "Synapse diff should be NEW - OLD")
	})
	t.Run("diffing large unrelated networks works", func(t *testing.T) {
		net1 := NewNetwork()
		net2 := NewNetwork()

		net1.Grow(4000, 10, 1000)
		net2.Grow(4000, 10, 1000)

		diff := DiffNetworks(net1, net2)

		ApplyDiff(diff, net1)
		ok, report := CheckIntegrity(net1)
		assert.True(t, ok, "merging unrelated networks failed integrity check", report)
	})
}

func Test_ApplyDiff(t *testing.T) {
	t.Run("synapse millivolts and activation history are properly applied", func(t *testing.T) {
		original := NewNetwork()
		syn1 := NewSynapse(original)
		syn1.Millivolts = 7

		// must link to a cell for integrity check
		cell := NewCell(original)
		syn1.FromNeuronAxon = cell.ID
		syn1.ToNeuronDendrite = cell.ID

		cloned := CloneNetwork(original)

		cloned.Synapses[syn1.ID].Millivolts = 14
		cloned.Synapses[syn1.ID].ActivationHistory = 42
		original.Synapses[syn1.ID].ActivationHistory = 3
		ok, report := CheckIntegrity(cloned)
		if !ok {
			cloned.Print()
		}
		assert.True(t, ok, "cloned network has bad integrity", report)

		diff := DiffNetworks(original, cloned)
		assert.Equal(t, 1, len(diff.synapseDiffs))
		assert.Equal(t, 1, len(diff.synapseFires))
		assert.Equal(t, int16(7), diff.synapseDiffs[syn1.ID])
		assert.Equal(t, uint(42), diff.synapseFires[syn1.ID])
		ApplyDiff(diff, original)

		assert.Equal(t, int16(14), original.Synapses[syn1.ID].Millivolts,
			"synapse millivolts failed to apply")
		assert.Equal(t, uint(45), original.Synapses[syn1.ID].ActivationHistory)
	})
	t.Run("all synapses and cells have valid connections after applying a diff", func(t *testing.T) {
		net1 := NewNetwork()
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
		assert.Equal(t, (50*5)+50+(200*10)+100, len(net1.Synapses))
		ok, report = CheckIntegrity(net1)
		assert.True(t, ok, report)
	})
	t.Run("adds new cell to the network", func(t *testing.T) {
	})

	t.Run("adds new synapse to the network", func(t *testing.T) {

	})
}

func Test_ApplyDiff_TrickeryIntegrityTests(t *testing.T) {
	t.Run("a cell ID is new and all synapses get reassigned, keeping network integrity", func(t *testing.T) {
		// network 1
		network := NewNetwork()
		n1Cell1 := NewCell(network)
		n1Cell2 := NewCell(network)

		network.linkCells(n1Cell1.ID, n1Cell2.ID)
		network.linkCells(n1Cell2.ID, n1Cell1.ID)

		ok, report := CheckIntegrity(network)
		assert.Equal(t, true, ok)
		if !ok {
			fmt.Println("network")
			report.Print()
		}

		// network 2
		net2 := NewNetwork()
		n2Cell1 := NewCell(net2)
		n2Cell2 := NewCell(net2)
		n2Cell3 := NewCell(net2)

		net2.linkCells(n2Cell1.ID, n2Cell2.ID)
		net2.linkCells(n2Cell2.ID, n2Cell1.ID)
		net2.linkCells(n2Cell3.ID, n2Cell1.ID)
		net2.linkCells(n2Cell2.ID, n2Cell3.ID)

		ok, report = CheckIntegrity(net2)
		assert.Equal(t, true, ok)
		if !ok {
			fmt.Println("net2")
			report.Print()
		}

		// time to test some things
		diff := DiffNetworks(network, net2)
		// diff.Print()
		assert.Equal(t, 1, len(diff.addedCells))
		assert.Equal(t, 2, len(diff.addedSynapses))

		ApplyDiff(diff, network)

		// network.Print()
		assert.Equal(t, 4, len(network.Synapses))
		assert.Equal(t, 3, len(network.Cells))

		ok, report = CheckIntegrity(network)
		assert.Equal(t, true, ok, "no integrity after apply diff")
	})
	t.Run("when a cell is new and another already exists one of its synapses was new", func(t *testing.T) {
		t.Run("the old synapse ID is removed from the dendrite synapse list", func(t *testing.T) {
			network := NewNetwork()
			assert.Equal(t, 0, len(network.Cells))
			assert.Equal(t, 0, len(network.Synapses))

			// create a new cell with a known ID
			// we will create a second network and add a cell on there
			// which purposely collides with one already on the network
			// we are merging onto
			cell := NewCell(network)
			receiver := NewCell(network)

			s1 := network.linkCells(cell.ID, receiver.ID)
			s1.Millivolts = 7

			pretestok, _ := CheckIntegrity(network)
			assert.Equal(t, true, pretestok)
			assert.Equal(t, 1, len(network.Synapses))
			assert.Equal(t, 2, len(network.Cells))

			net2 := NewNetwork()
			NewCell(net2)
			NewCell(net2) // a couple of extras
			cell2 := NewCell(net2)
			receiver2 := NewCell(net2)

			s2 := net2.linkCells(cell2.ID, receiver2.ID)
			s3 := net2.linkCells(receiver2.ID, cell2.ID)
			s2.Millivolts = 10
			s3.Millivolts = 11

			assert.Equal(t, 2, len(net2.Synapses))
			assert.Equal(t, 4, len(net2.Cells))
			pretestNet2ok, _ := CheckIntegrity(net2)
			assert.Equal(t, true, pretestNet2ok)

			// now do the diffing and checking
			diff := DiffNetworks(network, net2)
			assert.Equal(t, 0, len(diff.synapseDiffs))
			assert.Equal(t, 2, len(diff.addedCells))
			assert.Equal(t, 0, len(diff.synapseFires))
			assert.Equal(t, 2, len(diff.addedSynapses))
			// the diff is right. let's apply it.

			ApplyDiff(diff, network)
			assert.Equal(t, 4, len(network.Cells), "wrong number of cells on original network after diff applied")
			assert.Equal(t, 3, len(network.Synapses), "wrong number of synapses on original network after diff applied", network.Synapses, diff)

			postMergeIntegrityOK, report := CheckIntegrity(network)
			assert.Equal(t, true, postMergeIntegrityOK)
			if !postMergeIntegrityOK {
				report.Print()
			}
		})
	})
}

func Test_copyCellToNetwork(t *testing.T) {
	t.Run("copying a cell sets the pointer to a new network", func(t *testing.T) {
		originalNetwork := NewNetwork()
		newNetwork := NewNetwork()
		originalCell := NewCell(originalNetwork)
		copyCellToNetwork(originalCell, newNetwork)
		exists := newNetwork.CellExists(originalCell.ID)
		if !exists {
			t.Error("cell not copied to new network")
		}
		copiedCell := newNetwork.GetCell(originalCell.ID)
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

func Test_DiffPrint(t *testing.T) {
	t.Run("diff.Print works", func(t *testing.T) {
		net1 := NewNetwork()
		net2 := NewNetwork()
		net1.Grow(5, 2, 5)
		net2.Grow(5, 2, 5)
		diff := DiffNetworks(net1, net2)
		diff.Print()
	})
}
