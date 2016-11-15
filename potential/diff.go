package potential

/*
DiffNetworks produces a diff from the original network, showing the forward changes
from the newerNetwork. The diff is a map where the keys are synapse IDs, and the value
is the difference between the new and old network.

You can take the diff and apply it to the original network using addition,
by looping through the synapses and adding it.
*/
func DiffNetworks(originalNetwork, newerNetwork *Network) (diff map[int]int8) {
	return diff
}
