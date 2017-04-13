package cmd

import (
	"fmt"

	"github.com/ruffrey/nurtrace/potential"
)

// CompareFiringPatterns prints information about a network or requested components of the network.
func CompareFiringPatterns(network *potential.Network, cell1, cell2 []potential.CellID) (err error) {
	network.PrintTotals()
	fmt.Println("Cell group 1:", cell1)
	fmt.Println("Cell group 2:", cell2)
	network.ResetForTraining()
	patt1 := potential.FireNetworkUntilDone(network, cell1)
	network.ResetForTraining()
	patt2 := potential.FireNetworkUntilDone(network, cell2)
	fmt.Println("Cell 1 pattern:", patt1)
	fmt.Println("Cell 2 pattern:", patt2)
	diff := potential.DiffFiringPatterns(patt1, patt2)
	fmt.Println("Shared cells fired:", diff.Shared)
	fmt.Println("Unshared cells fired:", diff.Shared)
	fmt.Println("Ratio:", diff.Ratio())
	return nil
}
