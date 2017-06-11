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
		fmt.Println(a.WasFired, b.WasFired, c.WasFired)
		assert.Equal(t, 3, len(result), "wrong number of cells fired")
		fmt.Println("result=", result)
		assert.Equal(t, uint16(0), result[d.ID], "should not have fired this cell")
		assert.Equal(t, uint16(3), result[a.ID], "did not fire cell: a-0")
		assert.Equal(t, uint16(3), result[b.ID], "did not fire cell: b-1")
		assert.Equal(t, uint16(3), result[c.ID], "did not fire cell: c-2")
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
		r, _ := diff.Ratio()
		assert.Equal(t, 1.0, r)
	})
	t.Run("totally different firing patterns have ratio of 0", func(t *testing.T) {
		fp1 := make(FiringPattern)
		fp2 := make(FiringPattern)

		fp1[CellID(0)] = 1
		fp1[CellID(10)] = 2

		fp2[CellID(77)] = 3
		fp2[CellID(99)] = 4

		diff := DiffFiringPatterns(fp1, fp2)
		r, _ := diff.Ratio()
		assert.Equal(t, 0.0, r)
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
		r, _ := diff.Ratio()
		// (1 + 2 + 4) / (3 + 14 + 16 + 4)
		tot := 2 + 16 + 4.0
		assert.Equal(t, (tot-(1+2+4.0))/tot, r)
	})
}

func Test_FiringPatternMerge(t *testing.T) {
	t.Run("merging firing patterns returns a new combined pattern", func(t *testing.T) {
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

		merged := mergeFiringPatterns(fp1, fp2)

		assert.Equal(t, uint16(1), merged[CellID(0)])
		assert.Equal(t, uint16(15), merged[CellID(1)])
		assert.Equal(t, uint16(4), merged[CellID(2)])
	})
}
