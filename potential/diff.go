package potential

import (
	"fmt"
)

/*
Diff holds the changed values in the second network since the original network was cloned.

Diff values are always *old minus new*. They can be positive or negative. The diff would be
added back later when needed.
*/
type Diff struct {
	NetworkVersion string
	/*
	   synapses is a map where the keys are synapse IDs, and the value is the difference between
	   the new and old network on the `syanpse.Millivolts` property.
	*/
	synapseDiffs  map[SynapseID]int16
	synapseFires  map[SynapseID]uint
	addedSynapses []*Synapse
	addedCells    map[CellID]*Cell
}

/*
Print prints the diff to stdout
*/
func (diff *Diff) Print() {
	fmt.Println("Diff")
	fmt.Println("  synapseDiffs", diff.synapseDiffs)
	fmt.Println("  synapseFires", diff.synapseFires)
	fmt.Println("  addedSynapse")
	for _, s := range diff.addedSynapses {
		fmt.Println("    ", s.ID)
	}
	fmt.Println("  addedCells")
	for _, c := range diff.addedCells {
		fmt.Println("    ", c.ID)
	}
}

/*
NewDiff is a Diff factory
*/
func NewDiff() Diff {
	return Diff{
		synapseDiffs: make(map[SynapseID]int16),
		synapseFires: make(map[SynapseID]uint),
		addedCells:   make(map[CellID]*Cell),
	}
}

const zero int16 = 0

/*
DiffNetworks produces a diff from the original network, showing the forward changes
from the newerNetwork.

A synapse is the same if it:
- has the same cell ID
- has the same dendrite connection
- has the same axon connection

We cannot know if a cell is the same or not, but it does not create an integrity problem,
so that's fine. A cell is really just a collection of synapses.
*/
func DiffNetworks(originalNetwork, newerNetwork *Network) (diff Diff) {
	diff = NewDiff()

	// Get new synapses and the millivolt differences between existing synapses
	for id, newerNetworkSynapse := range newerNetwork.Synapses {
		alreadyExisted := originalNetwork.synExists(SynapseID(id))
		if !alreadyExisted {
			// we may want it derefenced so it is an instance, not a pointer. that will ensure
			// later we can need to update the synapse to be pointing to the originalNetwork
			// dendrite and axon cells. it will still be pointing to the old ones.
			diff.addedSynapses = append(diff.addedSynapses, newerNetworkSynapse)
			continue
		}
		originalSynapse := originalNetwork.getSyn(SynapseID(id))
		// the syanpse ID is not unique on the original network.
		// we need to make sure we didn't happen to generate it in a collision, though.
		isSame := newerNetworkSynapse.ToNeuronDendrite == originalSynapse.ToNeuronDendrite && newerNetworkSynapse.FromNeuronAxon == originalSynapse.FromNeuronAxon
		if !isSame {
			// tricked us! it generated the same random synapse ID, but the connections
			// to cells are different. treat it like a new synapse. During the copy
			// to the original network, its it will get regenerated to be unique.
			diff.addedSynapses = append(diff.addedSynapses, newerNetworkSynapse)
			continue
		}

		// this synapse already existed, so we will calculate the diff
		d := newerNetworkSynapse.Millivolts - originalSynapse.Millivolts
		if d != zero {
			diff.synapseDiffs[SynapseID(id)] = d
		}

		// Track how many times it fired, so when many training sessions are in play
		// we know if a cell should be considered for pruning. It is not a diff but will be
		// added later.
		if newerNetworkSynapse.ActivationHistory != 0 {
			diff.synapseFires[SynapseID(id)] = newerNetworkSynapse.ActivationHistory
		}
	}

	// Get new cells that were added to the network
	for id, newerNetworkCell := range newerNetwork.Cells {
		alreadyExisted := originalNetwork.cellExists(CellID(id))
		if !alreadyExisted {
			diff.addedCells[CellID(id)] = newerNetworkCell
		}
		// Here, we could theoretically get diff information on existing cells.
		// However, that *should* be captured by the synapses, and applying the diff
		// would add the additional synapses, and remove the gone synapses, which
		// are stored secondarily in each Cell.
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
	// New cells
	for _, cell := range diff.addedCells {
		copyCellToNetwork(cell, originalNetwork)
	}

	// New synapses
	for _, synapse := range diff.addedSynapses {

		// If this synapse was attached to a new cell, we need
		// to delete that cell's old synapse reference because the cell
		// would be is connected to the newer network's synapse now.
		// This was a long standing bug. It only causes a problem
		// after the next prune. We can end up with cells where a synapse
		// here or there is missing after the prune. Presumably because
		// there was always a "correct" but invalid ID reference.
		copySynapseToNetwork(synapse, originalNetwork)
		// old axon connection removed from cell if cell is new
		if _, isNewCell := diff.addedCells[synapse.FromNeuronAxon]; isNewCell {
			// fmt.Println("  removing old synapse reference from new cell (axon)", synapse.FromNeuronAxon)

			delete(originalNetwork.Cells[synapse.FromNeuronAxon].AxonSynapses, synapse.ID)
		}

		// old dendrite connection removed from cell if cell is new
		if _, isNewCell := diff.addedCells[synapse.ToNeuronDendrite]; isNewCell {
			// fmt.Println("  removing old synapse reference from new cell (dendrite)", synapse.ToNeuronDendrite)

			delete(originalNetwork.Cells[synapse.ToNeuronDendrite].DendriteSynapses, synapse.ID)
		}
	}

	// Update voltages and activations on existing synapses
	for synapseID, diffValue := range diff.synapseDiffs {
		s := originalNetwork.getSyn(synapseID)
		s.Millivolts += diffValue
	}
	// add the activation history
	for synapseID, activations := range diff.synapseFires {
		s := originalNetwork.getSyn(synapseID)
		s.ActivationHistory += activations
	}

	return nil
}

/*
CloneNetwork returns an exact copy of a network - not a pointer. This is useful when
doing distributed training.

It involves resetting pointers.
*/
func CloneNetwork(originalNetwork *Network) *Network {
	newNetwork := NewNetwork()

	originalNetwork.cellMux.Lock()
	for _, cell := range originalNetwork.Cells {
		originalNetwork.cellMux.Unlock()
		copyCellToNetwork(cell, newNetwork)
		originalNetwork.cellMux.Lock()
	}
	originalNetwork.cellMux.Unlock()
	originalNetwork.synMux.Lock()
	for _, synapse := range originalNetwork.Synapses {
		originalNetwork.synMux.Unlock()
		copySynapseToNetwork(synapse, newNetwork)
		originalNetwork.synMux.Lock()
	}
	originalNetwork.synMux.Unlock()

	return newNetwork
}

/*
copyCellToNetwork copies the properies of once cell to a new one, and updates the network pointer
on the new cell to a different given network. It also adds the cell to the new network.

Returns the cell ID of the copied cell, in case it had to change.
*/
func copyCellToNetwork(origCell *Cell, newNetwork *Network) {
	// fresh cell on a new network that will serve as the copy
	copiedCell := NewCell(newNetwork)

	copiedCell.Network = newNetwork

	copiedCell.Immortal = origCell.Immortal
	copiedCell.activating = origCell.activating
	copiedCell.Voltage = origCell.Voltage

	// golang does not copy a map on assignment; must loop over it.

	newNetwork.cellMux.Lock()
	for synapseID := range origCell.AxonSynapses {
		copiedCell.AxonSynapses[synapseID] = true
	}
	for synapseID := range origCell.DendriteSynapses {
		copiedCell.DendriteSynapses[synapseID] = true
	}
	newNetwork.cellMux.Unlock()
}

/*
copySynapseToNetwork copies the properies of once synapse to a new one, and updates the network pointer
on the new synapse to a different given network.

Returns its synapse ID because it had to change.
*/
func copySynapseToNetwork(synapse *Synapse, newNetwork *Network) SynapseID {
	// lock old and new networks to prevent map reads while we are adding
	// or removing map values!

	// The new synapse will have a different ID, because it will surely
	// have a different position in the array.
	copiedSynapse := NewSynapse(newNetwork)

	// copy properties to the new synapse

	// pre and post synaptic neurons
	copiedSynapse.FromNeuronAxon = synapse.FromNeuronAxon
	copiedSynapse.ToNeuronDendrite = synapse.ToNeuronDendrite

	copiedSynapse.Millivolts = synapse.Millivolts
	// we do need to keep this property because we will want to
	// grow/prune the synapse later
	copiedSynapse.ActivationHistory = synapse.ActivationHistory

	// Make sure that this synapse's cells have an ID reffing it.
	// If there happened to also be a cell ID collision for its axon or dendrite,
	// the connection on the cell to this synapse would be missing.

	d := copiedSynapse.ToNeuronDendrite
	a := copiedSynapse.FromNeuronAxon
	dCell := newNetwork.getCell(d)
	aCell := newNetwork.getCell(a)

	newNetwork.cellMux.Lock()
	dCell.DendriteSynapses[copiedSynapse.ID] = true
	aCell.AxonSynapses[copiedSynapse.ID] = true
	newNetwork.cellMux.Unlock()

	return copiedSynapse.ID
}
