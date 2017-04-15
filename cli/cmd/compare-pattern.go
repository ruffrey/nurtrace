package cmd

import (
	"fmt"

	"github.com/ruffrey/nurtrace/potential"
)

// CompareFiringPatterns prints information about a network or requested components of the network.
func CompareFiringPatterns(network *potential.Network, cell1, cell2 map[potential.CellID]bool, n int) (err error) {
	network.PrintTotals()
	fmt.Println("Cell group 1:", cell1)
	fmt.Println("Cell group 2:", cell2)
	fmt.Println("Firing", n, "times")
	network.ResetForTraining()
	var patt1 potential.FiringPattern
	var patt2 potential.FiringPattern

	for i := 0; i < n; i++ {
		isLast := i == (n - 1)
		p := potential.FireNetworkUntilDone(network, cell1)
		if isLast {
			patt1 = p
		}
	}
	network.ResetForTraining()
	for i := 0; i < n; i++ {
		isLast := i == (n - 1)
		p := potential.FireNetworkUntilDone(network, cell2)
		if isLast {
			patt2 = p
		}
	}

	fmt.Println("Cell 1 pattern:", patt1)
	fmt.Println("Cell 2 pattern:", patt2)
	diff := potential.DiffFiringPatterns(patt1, patt2)
	fmt.Println("Shared cells fired:", diff.Shared)
	fmt.Println("Unshared cells fired:", diff.Unshared)
	fmt.Println("Ratio:", diff.Ratio())
	return nil
}
