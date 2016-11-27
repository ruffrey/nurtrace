package potential

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewCell(t *testing.T) {
	t.Run("calling NewCell() also adds it to the network", func(t *testing.T) {
		network := NewNetwork()
		cell := NewCell(&network)
		netCell, ok := network.Cells[cell.ID]
		if !ok {
			panic("NewCell() did not add cell to the network")
		}
		assert.Equal(t, cell.ID, netCell.ID, "NewCell() added cell to network but ids do not match")
	})
}
