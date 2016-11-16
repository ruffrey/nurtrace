package potential

import "github.com/jinzhu/copier"

/*
Diff holds the changed values in the second network since the original network was cloned.
*/
type Diff struct {
	/*
	   synapses is a map where the keys are synapse IDs, and the value is the difference between
	   the new and old network.
	*/
	synapseDiffs    map[int]int8
	addedSynapses   []Synapse
	removedSynapses []Synapse
	/*
	   For cells that were kept, this is the voltage change from old to new. They would be
	   added to the old network (though can be positive or negative). The key is the cell ID.
	*/
	cellVoltageDiffs map[int]int8
	removedCells     []Cell
	addedCells       []Cell
}

/*
DiffNetworks produces a diff from the original network, showing the forward changes
from the newerNetwork.

You can take the diff and apply it to the original network using addition,
by looping through the synapses and adding it.
*/
func DiffNetworks(originalNetwork, newerNetwork *Network) (diff Diff) {
	return diff
}

/*
ApplyDiff uses a diff between the originalNetwork and another duplicate network.

It updates (CHANGES) the originalNetwork using the synapse weight changes from the diff.

The originalNetwork should probably be in a resting state when the diff is applied,
but this isn't technically required. Though, it is undefined behavior.
*/
func ApplyDiff(diff Diff, originalNetwork *Network) {
	for synapseID, diffValue := range diff.synapseDiffs {
		originalNetwork.Synapses[synapseID].Millivolts += diffValue
	}
	// TODO: more
}

/*
CloneNetwork returns an exact copy of a network - not a pointer. This is useful when
doing distributed testing.
*/
func CloneNetwork(network *Network) Network {
	newNetwork := NewNetwork()
	copier.Copy(&newNetwork, network)
	return newNetwork
}
