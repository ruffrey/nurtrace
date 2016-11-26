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
		cell := NewCell(&network)
		synapse := NewSynapse(&network)

		// tests are here
		diff.addedCells = append(diff.addedCells, &cell)
		diff.addedSynapses = append(diff.addedSynapses, &synapse)
		diff.removedCells = append(diff.removedCells, cell.ID)
		diff.removedSynapses = append(diff.removedSynapses, synapse.ID)
	})
}

func Test_CopyNetwork(t *testing.T) {
	original := NewNetwork()
	beforeCell := NewCell(&original)
	original.Cells[beforeCell.ID] = &beforeCell
	cloned := CloneNetwork(&original)
	// change something
	cloned.SynapseLearnRate = 100

	if cloned.SynapseLearnRate == original.SynapseLearnRate {
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
}

func Test_ApplyDiff(t *testing.T) {
	t.Run("synapse millivolts are properly applied", func(t *testing.T) {
		original := NewNetwork()
		syn1 := NewSynapse(&original)
		fmt.Println(syn1)

		original.Synapses[syn1.ID] = &syn1
		cloned := CloneNetwork(&original)

		diff, err := DiffNetworks(&original, &cloned)

		assert.Equal(t, err, nil, "Expected no error during diff")
		fmt.Println("diff=", diff)
	})
}
