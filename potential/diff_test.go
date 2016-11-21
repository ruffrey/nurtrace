package potential

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DiffNetworks(t *testing.T) {

}

func Test_CopyNetwork(t *testing.T) {
	original := NewNetwork()
	beforeCell := NewCell()
	original.Cells[beforeCell.ID] = &beforeCell
	cloned := CloneNetwork(&original)
	afterCell := original.Cells[0]
	// change something
	cloned.SynapseLearnRate = 100

	assert.EqualValues(t, beforeCell.ID, afterCell.ID, "before and after cell IDs should be equal")
	assert.NotEqual(t, original.SynapseLearnRate, cloned.SynapseLearnRate,
		"cloned and original should not point to the same network")
}
