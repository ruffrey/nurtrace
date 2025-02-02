package cmd

import (
	"log"

	"github.com/ruffrey/nurtrace/potential"
)

// CompareFiringPatterns prints information about a network or requested components of the network.
func CompareFiringPatterns(network *potential.Network, cell1, cell2 potential.FiringPattern, n int) (err error) {
	network.PrintTotals()
	log.Println("Cell group 1:", cell1)
	log.Println("Cell group 2:", cell2)
	log.Println("Firing", n, "times")
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

	log.Println("Cell 1 pattern:", patt1)
	log.Println("Cell 2 pattern:", patt2)
	log.Println("Ratio:", potential.DiffFiringPatterns(patt1, patt2))
	return nil
}
