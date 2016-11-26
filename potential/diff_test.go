package potential

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func Test_DiffNetworks(t *testing.T) {
	// t.Run("synapse millivolts are properly applied", func(t *testing.T) {
	// 	original := NewNetwork()
	// 	beforeCell := NewCell(&original)
	// 	syn1 := NewSynapse(&original)
	//
	// 	original.Cells[beforeCell.ID] = &beforeCell
	// 	cloned := CloneNetwork(&original)
	// 	afterCell := original.Cells[0]
	//
	// })
}
