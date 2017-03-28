package cmd

import "github.com/ruffrey/nurtrace/potential"

// Merge merges two networks and also returns the diff
func Merge(originalNetworkFilename, otherNetworkFilename string) (originalNetwork *potential.Network, diff potential.Diff, err error) {
	originalNetwork, err = potential.LoadNetworkFromFile(originalNetworkFilename)
	if err != nil {
		return originalNetwork, diff, err
	}
	otherNetwork, err := potential.LoadNetworkFromFile(otherNetworkFilename)
	if err != nil {
		return originalNetwork, diff, err
	}
	diff = potential.DiffNetworks(originalNetwork, otherNetwork)

	potential.ApplyDiff(diff, originalNetwork)

	return originalNetwork, diff, nil
}
