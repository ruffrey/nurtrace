package potential

import "fmt"

/*
Diff holds the changed values in the second network since the original network was cloned.

Diff values are always *old minus new*. They can be positive or negative. The diff would be
added back later when needed.
*/
type Diff struct {
	NetworkVersion Shasum
	/*
	   synapses is a map where the keys are synapse IDs, and the value is the difference between
	   the new and old network.
	*/
	synapseDiffs  map[SynapseID]int8
	addedSynapses []*Synapse
	/*
	   removedSynapses is a list of the IDs of the synapses that no longer exist in the new network.
	*/
	removedSynapses []SynapseID
	addedCells      []*Cell
	/*
	   removedCells is a list of the cell IDs that no longer exist in the new network.
	*/
	removedCells []CellID
}

/*
NewDiff is a Diff factory
*/
func NewDiff() Diff {
	return Diff{
		synapseDiffs: make(map[SynapseID]int8),
	}
}

/*
DiffNetworks produces a diff from the original network, showing the forward changes
from the newerNetwork.

You can take the diff and apply it to the original network using addition,
by looping through the synapses and adding it.
*/
func DiffNetworks(originalNetwork, newerNetwork *Network) (diff Diff) {
	diff = NewDiff()
	diff.NetworkVersion = originalNetwork.Version

	// Get new synapses and the millivolt differences between existing synapses
	for id, newerNetworkSynapse := range newerNetwork.Synapses {
		originalSynapse, alreadyExisted := originalNetwork.Synapses[id]
		if !alreadyExisted {
			// we may want it derefenced so it is an instance, not a pointer. that will ensure
			// later we can need to update the synapse to be pointing to the originalNetwork
			// dendrite and axon cells. it will still be pointing to the old ones.
			diff.addedSynapses = append(diff.addedSynapses, newerNetworkSynapse)
		} else {
			// this synapse already existed, so we will calculate the diff
			diff.synapseDiffs[id] = newerNetworkSynapse.Millivolts - originalSynapse.Millivolts
		}
	}
	// Check if any synapses were removed
	for id, originalNetworkSynapse := range originalNetwork.Synapses {
		_, stillExists := newerNetwork.Synapses[id]
		if !stillExists {
			diff.removedSynapses = append(diff.removedSynapses, originalNetworkSynapse.ID)
		}
	}
	// Get new cells that were added to the network
	for id, newerNetworkCell := range newerNetwork.Cells {
		_, alreadyExisted := originalNetwork.Cells[id]
		if !alreadyExisted {
			diff.addedCells = append(diff.addedCells, newerNetworkCell)
		} else {
			// Here, we could theoretically get diff information on existing cells.
			// However, that *should* be captured by the synapses, and applying the diff
			// would add the additional synapses, and remove the gone synapses, which
			// are stored secondarily in each Cell.
		}
	}
	// Check if any cells were removed
	for id, originalNetworkCell := range originalNetwork.Cells {
		_, stillExists := newerNetwork.Cells[id]
		if !stillExists {
			diff.removedCells = append(diff.removedCells, originalNetworkCell.ID)
		}
	}

	return diff
}

/*
ApplyDiff uses a diff between the originalNetwork and another duplicate network.

It updates (CHANGES) the originalNetwork using the synapse weight changes from the diff.

The originalNetwork should probably be in a resting state when the diff is applied,
but this isn't technically required. Though, it is undefined behavior if not.

When training a network, we should be starting all cells at resting potential voltage. So
no need to copy any changes from existing cell voltages.

**The order of operations in ApplyDiff matters!**

This will update the network version, also.
*/
func ApplyDiff(diff Diff, originalNetwork *Network) (err error) {
	// Sanity check
	if originalNetwork.Version != diff.NetworkVersion {
		err = fmt.Errorf("Diffing networks failed due to version mismatch: original=%s, newer=%s", originalNetwork.Version, diff.NetworkVersion)
		return err
	}

	// New cells
	for _, cell := range diff.addedCells {
		copyCellToNetwork(cell, originalNetwork)
	}

	// New synapses
	for _, synapse := range diff.addedSynapses {
		copySynapseToNetwork(synapse, originalNetwork)
		// add connections to cells
		originalNetwork.Cells[synapse.FromNeuronAxon].AxonSynapses[synapse.ID] = true
		originalNetwork.Cells[synapse.ToNeuronDendrite].DendriteSynapses[synapse.ID] = true
	}

	// Update voltages on existing synapses
	for synapseID, diffValue := range diff.synapseDiffs {
		synapse := originalNetwork.Synapses[synapseID]
		synapse.Millivolts += diffValue
	}

	// Remove synapses
	for _, synapseID := range diff.removedSynapses {
		// Removes synapses from both connected cells.
		// Will also prune cells with no synapses.
		originalNetwork.PruneSynapse(synapseID)
	}

	// Remove cells by ID
	for _, cellID := range diff.removedCells {
		delete(originalNetwork.Cells, cellID)
	}

	originalNetwork.RegenVersion()

	return nil
}

/*
CloneNetwork returns an exact copy of a network - not a pointer. This is useful when
doing distributed training.

It involves resetting pointers.
*/
func CloneNetwork(originalNetwork *Network) *Network {
	n := NewNetwork()
	newNetwork := &n
	newNetwork.SynapseLearnRate = originalNetwork.SynapseLearnRate
	newNetwork.SynapseMinFireThreshold = originalNetwork.SynapseMinFireThreshold

	for _, cell := range originalNetwork.Cells {
		copyCellToNetwork(cell, newNetwork)
	}
	for _, synapse := range originalNetwork.Synapses {
		copySynapseToNetwork(synapse, newNetwork)
	}

	return newNetwork
}

/*
copyCellToNetwork copies the properies of once cell to a new one, and updates the network pointer
on the new cell to a different given network. It also adds the cell to the new network.
*/
func copyCellToNetwork(cell *Cell, newNetwork *Network) {
	copiedCell := NewCell()
	copiedCell.ID = cell.ID
	newNetwork.Cells[cell.ID] = copiedCell
	copiedCell.Network = newNetwork

	copiedCell.Immortal = cell.Immortal
	copiedCell.activating = cell.activating
	copiedCell.Voltage = cell.Voltage
	copiedCell.AxonSynapses = cell.AxonSynapses
	copiedCell.DendriteSynapses = cell.DendriteSynapses
}

/*
copySynapseToNetwork copies the properies of once synapse to a new one, and updates the network pointer
on the new synapse to a different given network.
*/
func copySynapseToNetwork(synapse *Synapse, newNetwork *Network) {
	copiedSynapse := NewSynapse()
	copiedSynapse.ID = synapse.ID
	newNetwork.Synapses[synapse.ID] = copiedSynapse
	copiedSynapse.Network = newNetwork

	copiedSynapse.Millivolts = synapse.Millivolts
	copiedSynapse.FromNeuronAxon = synapse.FromNeuronAxon
	copiedSynapse.ToNeuronDendrite = synapse.ToNeuronDendrite
	// we do need to keep this because we might want to grow the synapse later
	copiedSynapse.ActivationHistory = synapse.ActivationHistory
}
