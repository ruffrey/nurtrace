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
