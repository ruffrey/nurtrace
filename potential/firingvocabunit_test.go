package potential

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_VocabUnit(t *testing.T) {
	t.Run("isCellOnAnyInput returns true when cell is on inputs", func(t *testing.T) {
		inputs := make(map[InputValue]*VocabUnit)
		cellID := CellID(5)

		g := InputValue("g")

		inputs[g] = &VocabUnit{
			Value:      g,
			InputCells: make(FiringPattern),
		}

		inputs[g].InputCells[cellID] = 3 // the one in question
		inputs[g].InputCells[CellID(2)] = 2
		inputs[g].InputCells[CellID(13)] = 1

		assert.True(t, isCellOnAnyInput(cellID, inputs))
	})
	t.Run("isCellOnAnyInput returns false when cell is NOT on inputs", func(t *testing.T) {
		inputs := make(map[InputValue]*VocabUnit)
		cellID := CellID(5)

		g := InputValue("g")

		inputs[g] = &VocabUnit{
			Value:      g,
			InputCells: make(FiringPattern),
		}

		// cellID is not here
		inputs[g].InputCells[CellID(6)] = 2
		inputs[g].InputCells[CellID(18)] = 1

		assert.False(t, isCellOnAnyInput(cellID, inputs))
	})
}
