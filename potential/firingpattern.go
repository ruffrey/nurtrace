package potential

import (
	"github.com/ruffrey/nurtrace/laws"
)

// FiringPattern represents all cells that fired in a single step
type FiringPattern map[CellID]bool

/*
fireNetworkUntilDone takes some seed cells then fires the network until
it has no more firing, up to `laws.MaxPostFireSteps`.
*/
func fireNetworkUntilDone(network *Network, seedCells []CellID) (fp FiringPattern) {
	var i uint8
	fp = make(map[CellID]bool)
	for _, cellID := range seedCells {
		network.getCell(cellID).FireActionPotential()
		network.resetCellsOnNextStep[cellID] = true
	}
	// we ignore the seedCells
	for {
		if i >= laws.MaxPostFireSteps {
			break
		}
		hasMore := network.Step()
		for cellID := range network.resetCellsOnNextStep {
			fp[cellID] = true
		}
		if !hasMore {
			break
		}
		i++
	}
	return fp
}
