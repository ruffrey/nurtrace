package potential

import (
	"fmt"
	"testing"

	"github.com/ruffrey/nurtrace/laws"
	"github.com/stretchr/testify/assert"
)

func Test_FiringPattern(t *testing.T) {
	t.Run("FireNetworkUntilDone fires the seed cells and returns the fired ones", func(t *testing.T) {
		network := NewNetwork()
		a := NewCell(network)
		a.Tag = "a"
		b := NewCell(network)
		b.Tag = "b"
		c := NewCell(network)
		c.Tag = "c"
		d := NewCell(network)
		d.Tag = "d"

		// purposely in a forever firing loop to make sure it exits

		s1 := network.linkCells(a.ID, b.ID)
		s2 := network.linkCells(b.ID, c.ID)
		s3 := network.linkCells(c.ID, a.ID)
		network.linkCells(d.ID, c.ID) // will not fire
		s1.Millivolts = laws.ActualSynapseMax
		s2.Millivolts = laws.ActualSynapseMax
		s3.Millivolts = laws.ActualSynapseMax

		// never fires d
		cells := make(FiringPattern)
		cells[a.ID] = 1
		result := FireNetworkUntilDone(network, cells)
		fmt.Println(a.activating, b.activating, c.activating)
		assert.Equal(t, 3, len(result), "wrong number of cells fired")
		assert.Equal(t, false, result[d.ID], "should not have fired this cell")
		assert.Equal(t, true, result[a.ID], "did not fire cell: a")
		assert.Equal(t, true, result[b.ID], "did not fire cell: b")
		assert.Equal(t, true, result[c.ID], "did not fire cell: c")
	})
}

func Test_FiringDiffRatio(t *testing.T) {
	t.Run("identical firing patterns have ratio of 1", func(t *testing.T) {
		fp1 := make(FiringPattern)
		fp2 := make(FiringPattern)

		// 1 different
		fp1[CellID(0)] = 2
		fp2[CellID(0)] = 2

		// 2 different
		fp1[CellID(1)] = 7
		fp2[CellID(1)] = 7

		diff := DiffFiringPatterns(fp1, fp2)
		assert.Equal(t, 1.0, diff)
	})
	t.Run("totally different firing patterns have ratio of 0", func(t *testing.T) {
		fp1 := make(FiringPattern)
		fp2 := make(FiringPattern)

		fp1[CellID(0)] = 1
		fp1[CellID(10)] = 2

		fp2[CellID(77)] = 3
		fp2[CellID(99)] = 4

		diff := DiffFiringPatterns(fp1, fp2)
		assert.Equal(t, 0.0, diff)
	})
	t.Run("radio calculates the number of unrepresented fires to represented fires", func(t *testing.T) {
		fp1 := make(FiringPattern)
		fp2 := make(FiringPattern)

		// 1 different
		fp1[CellID(0)] = 2
		fp2[CellID(0)] = 1

		// 2 different
		fp1[CellID(1)] = 14
		fp2[CellID(1)] = 16

		// 4 different / unshared
		fp1[CellID(2)] = 4

		diff := DiffFiringPatterns(fp1, fp2)

		// (1 + 2 + 4) / (3 + 14 + 16 + 4)
		tot := 2 + 16 + 4.0
		assert.Equal(t, (tot-(1+2+4.0))/tot, diff)
	})
}
