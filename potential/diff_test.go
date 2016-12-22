package potential

import (
	"fmt"
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
		diff.addedCells[cell.ID] = cell
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

		assert.ObjectsAreEqualValues(original.Cells, cloned.Cells)
		assert.ObjectsAreEqualValues(original.Synapses, cloned.Synapses)
		assert.Equal(t, len(original.Cells), len(cloned.Cells))
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
	t.Run("diffing large unrelated networks works", func(t *testing.T) {
		n1 := NewNetwork()
		n2 := NewNetwork()
		net1 := &n1
		net2 := &n2

		n1.Grow(4000, 10, 1000)
		n2.Grow(4000, 10, 1000)

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
			cloned.Print()
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
	t.Run("adds new synapse to the network", func(t *testing.T) {

	})
}

func Test_ApplyDiff_TrickeryIntegrityTests(t *testing.T) {
	t.Run("a cell ID is new and one synapse ID is not unique so it gets reassigned, keeps network integrity", func(t *testing.T) {
		// network 1
		n := NewNetwork()
		network := &n
		n1Cell1 := NewCell(network)
		n1Cell2 := NewCell(network)
		delete(network.Cells, n1Cell1.ID)
		delete(network.Cells, n1Cell2.ID)
		n1Cell1.ID = 11
		n1Cell2.ID = 12
		network.Cells[n1Cell1.ID] = n1Cell1
		network.Cells[n1Cell2.ID] = n1Cell2
		n1Syn1 := NewSynapse(network)
		n1Syn2 := NewSynapse(network)
		delete(network.Synapses, n1Syn1.ID)
		delete(network.Synapses, n1Syn2.ID)
		n1Syn1.ID = 101
		n1Syn2.ID = 102
		network.Synapses[n1Syn1.ID] = n1Syn1
		network.Synapses[n1Syn2.ID] = n1Syn2

		n1Syn1.FromNeuronAxon = n1Cell1.ID
		n1Syn1.ToNeuronDendrite = n1Cell2.ID
		n1Cell1.AxonSynapses[n1Syn1.ID] = true
		n1Cell2.DendriteSynapses[n1Syn1.ID] = true

		n1Syn2.FromNeuronAxon = n1Cell2.ID
		n1Syn2.ToNeuronDendrite = n1Cell1.ID
		n1Cell1.DendriteSynapses[n1Syn2.ID] = true
		n1Cell2.AxonSynapses[n1Syn2.ID] = true

		ok, report := CheckIntegrity(network)
		assert.Equal(t, true, ok)
		if !ok {
			fmt.Println("network")
			report.Print()
		}

		// network 2
		n2 := NewNetwork()
		net2 := &n2
		n2Cell1 := NewCell(net2)
		n2Cell2 := NewCell(net2)
		delete(net2.Cells, n2Cell1.ID)
		delete(net2.Cells, n2Cell2.ID)
		n2Cell1.ID = 21
		n2Cell2.ID = 12 // cell ID collision with first network
		net2.Cells[n2Cell1.ID] = n2Cell1
		net2.Cells[n2Cell2.ID] = n2Cell2
		n2Syn1 := NewSynapse(net2)
		n2Syn2 := NewSynapse(net2)
		delete(net2.Synapses, n2Syn1.ID)
		delete(net2.Synapses, n2Syn2.ID)
		n2Syn1.ID = 101 // synapse ID collision with first network
		n2Syn2.ID = 26  // synapse ID collision with first network
		net2.Synapses[n2Syn1.ID] = n2Syn1
		net2.Synapses[n2Syn2.ID] = n2Syn2

		n2Syn1.FromNeuronAxon = n2Cell1.ID
		n2Syn1.ToNeuronDendrite = n2Cell2.ID
		n2Cell1.AxonSynapses[n2Syn1.ID] = true
		n2Cell2.DendriteSynapses[n2Syn1.ID] = true

		n2Syn2.FromNeuronAxon = n2Cell2.ID
		n2Syn2.ToNeuronDendrite = n2Cell1.ID
		n2Cell1.DendriteSynapses[n2Syn2.ID] = true
		n2Cell2.AxonSynapses[n2Syn2.ID] = true

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

		t.Run("and can withstand pruning some things yet keep network integrity", func(t *testing.T) {
			network.Synapses[26].ActivationHistory = 100
			network.Prune()
			assert.Equal(t, 1, len(network.Synapses))
			assert.Equal(t, 2, len(network.Cells))
			ok, report = CheckIntegrity(network)
			assert.Equal(t, true, ok, "no integrity after apply diff")
			if !ok {
				report.Print()
			}
		})
	})
	t.Skip("when a cell is new and another already exists one of its synapses has been re-ID-d due to collision", func(t *testing.T) {
		t.Run("the old synapse ID is removed from the dendrite synapse list", func(t *testing.T) {
			n := NewNetwork()
			network := &n
			assert.Equal(t, 0, len(network.Cells))
			assert.Equal(t, 0, len(network.Synapses))

			// create a new cell with a known ID
			// we will create a second network and add a cell on there
			// which purposely collides with one already on the network
			// we are merging onto
			cell := NewCell(network)
			delete(network.Cells, cell.ID)
			cell.ID = 1
			network.Cells[cell.ID] = cell
			s1 := NewSynapse(network)
			s1.Millivolts = 7
			delete(network.Synapses, s1.ID)
			s1.ID = 101
			network.Synapses[s1.ID] = s1
			receiver := NewCell(network)
			delete(network.Cells, receiver.ID)
			receiver.ID = 2
			network.Cells[receiver.ID] = receiver
			s1.FromNeuronAxon = cell.ID
			cell.AxonSynapses[s1.ID] = true
			s1.ToNeuronDendrite = receiver.ID
			receiver.DendriteSynapses[s1.ID] = true
			pretestok, _ := CheckIntegrity(network)
			assert.Equal(t, true, pretestok)
			assert.Equal(t, 1, len(network.Synapses))
			assert.Equal(t, 2, len(network.Cells))
			// fmt.Println("network before")
			// network.Print()

			n2 := NewNetwork()
			net2 := &n2
			s2 := NewSynapse(net2)
			s2.Millivolts = 10
			delete(net2.Synapses, s2.ID)
			assert.Equal(t, 0, len(net2.Synapses))
			s2.ID = 101
			net2.Synapses[s2.ID] = s2
			cell2 := NewCell(net2)
			delete(net2.Cells, cell2.ID)
			cell2.ID = 1
			net2.Cells[cell2.ID] = cell2
			receiver2 := NewCell(net2)
			delete(net2.Cells, receiver2.ID)
			receiver2.ID = 44
			net2.Cells[receiver2.ID] = receiver2
			s2.FromNeuronAxon = cell2.ID
			s2.ToNeuronDendrite = receiver2.ID
			cell2.AxonSynapses[s2.ID] = true
			receiver2.DendriteSynapses[s2.ID] = true
			assert.Equal(t, 1, len(net2.Synapses))
			assert.Equal(t, 2, len(net2.Cells))
			pretestNet2ok, _ := CheckIntegrity(net2)
			assert.Equal(t, true, pretestNet2ok)
			// fmt.Println("net2 before")
			// net2.Print()

			// now do the diffing and checking
			diff := DiffNetworks(network, net2)
			// fmt.Println("diff:")
			// diff.Print()
			assert.Equal(t, 0, len(diff.synapseDiffs))
			assert.Equal(t, 1, len(diff.addedCells))
			assert.Equal(t, 0, len(diff.synapseFires))
			assert.Equal(t, 1, len(diff.addedSynapses))
			assert.Equal(t, s2, diff.addedSynapses[0])
			// the diff is right. let's apply it.

			ApplyDiff(diff, network)
			assert.Equal(t, 2, len(network.Synapses), "wrong number of synapses on original network after diff applied", network.Synapses)

			postMergeIntegrityOK, report := CheckIntegrity(network)
			assert.Equal(t, true, postMergeIntegrityOK)
			if !postMergeIntegrityOK {
				report.Print()
			}
			// fmt.Println("END NETWORK")
			// network.Print()
		})
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
