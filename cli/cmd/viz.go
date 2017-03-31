package cmd

import "github.com/ruffrey/nurtrace/potential"

// Viz creates a visualization of the network
func Viz(networkFile string) (err error) {
	network, err := potential.LoadNetworkFromFile(networkFile)
	if err != nil {
		return err
	}
	network.PrintTotals()
	return nil
}
