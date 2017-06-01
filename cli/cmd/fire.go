package cmd

import (
	"fmt"

	"github.com/ruffrey/nurtrace/potential"
)

// FireCell fires a cell n times and prints all cells that it fires.
func FireCell(network *potential.Network, cell potential.CellID, n int) (err error) {
	network.ResetForTraining()
	cellArg := make(potential.FiringPattern)
	cellArg[cell] = 1
	for i := 0; i < n; i++ {
		isLast := i == n-1
		if isLast {
			pattern := potential.FireNetworkUntilDone(network, cellArg)
			for firedCell := range pattern {
				fmt.Println(firedCell)
			}
			continue
		}
		potential.FireNetworkUntilDone(network, cellArg)
	}
	return nil
}
