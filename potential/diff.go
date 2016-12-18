package potential

import (
	"fmt"

	"github.com/y0ssar1an/q"
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
	synapseDiffs  map[SynapseID]int8
	synapseFires  map[SynapseID]uint
	addedSynapses []*Synapse
	addedCells    []*Cell
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
		synapseDiffs: make(map[SynapseID]int8),
		synapseFires: make(map[SynapseID]uint),
	}
}

const zero int8 = 0

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
		originalSynapse, alreadyExisted := originalNetwork.Synapses[id]
		if !alreadyExisted {
			// we may want it derefenced so it is an instance, not a pointer. that will ensure
			// later we can need to update the synapse to be pointing to the originalNetwork
			// dendrite and axon cells. it will still be pointing to the old ones.
			diff.addedSynapses = append(diff.addedSynapses, newerNetworkSynapse)
			continue
		}

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
			diff.synapseDiffs[id] = d
		}

		// Track how many times it fired, so when many training sessions are in play
		// we know if a cell should be considered for pruning. It is not a diff but will be
		// added later.
		if newerNetworkSynapse.ActivationHistory != 0 {
			diff.synapseFires[id] = newerNetworkSynapse.ActivationHistory
		}
	}

	// Get new cells that were added to the network
	for id, newerNetworkCell := range newerNetwork.Cells {
		_, alreadyExisted := originalNetwork.Cells[id]
		if !alreadyExisted {
			diff.addedCells = append(diff.addedCells, newerNetworkCell)
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
	if ok, report := CheckIntegrity(originalNetwork); !ok {
		fmt.Println("ApplyDiff: originalNetwork has no integrity BEFORE")
		report.Print()
		panic("no integrity")
	}
	// New cells
	cellIDChanges := make(map[CellID]CellID) // old:new
	for _, cell := range diff.addedCells {
		newID := copyCellToNetwork(cell, originalNetwork)
		if newID != cell.ID {
			fmt.Println("cell ID did change", cell.ID, newID)
			cellIDChanges[cell.ID] = newID
		}
	}

	// New synapses
	for _, synapse := range diff.addedSynapses {
		// If this synapse was attached to a cell where the cell.ID needed to change,
		// we need to point this synapse to it
		if newCellID, didChange := cellIDChanges[synapse.FromNeuronAxon]; didChange {
			fmt.Println("axon ID did change", synapse.FromNeuronAxon, newCellID)
			synapse.FromNeuronAxon = newCellID
		}
		if newCellID, didChange := cellIDChanges[synapse.ToNeuronDendrite]; didChange {
			fmt.Println("dendrite ID did change", synapse.ToNeuronDendrite, newCellID)
			synapse.ToNeuronDendrite = newCellID
		}

		// If the synapse already existed on the new network, and its id
		// got reassigned, we need to update this synapses cells to point to
		// the new synapse id.
		newSynapseID := copySynapseToNetwork(synapse, originalNetwork)
		if newSynapseID != synapse.ID {
			fmt.Println("synapse ID did change", synapse.ID, newSynapseID)
		}
		// add connections to cells
		newSyn, on := originalNetwork.Synapses[newSynapseID]
		if !on {
			fmt.Println("error: synapse ID was not copied to original network", newSynapseID)
			panic("synapse not copied to network during ApplyDiff")
		}
		originalNetwork.Cells[newSyn.ToNeuronDendrite].DendriteSynapses[newSynapseID] = true
		originalNetwork.Cells[newSyn.FromNeuronAxon].AxonSynapses[newSynapseID] = true
	}

	// Update voltages and activations on existing synapses
	for synapseID, diffValue := range diff.synapseDiffs {
		originalNetwork.Synapses[synapseID].Millivolts += diffValue
	}
	// add the activation history
	for synapseID, activations := range diff.synapseFires {
		originalNetwork.Synapses[synapseID].ActivationHistory += activations
	}

	if ok, report := CheckIntegrity(originalNetwork); !ok {
		fmt.Println("ApplyDiff: originalNetwork has no integrity AFTER")
		diff.Print()
		report.Print()
		panic("no integrity")
	}

	return nil
}

/*
CloneNetwork returns an exact copy of a network - not a pointer. This is useful when
doing distributed training.

It involves resetting pointers.
*/
func CloneNetwork(originalNetwork *Network) *Network {
	if ok, report := CheckIntegrity(originalNetwork); !ok {
		q.Q("CloneNetwork: originalNetwork has no integrity", report)
		panic("no integrity")
	}

	n := NewNetwork()
	newNetwork := &n

	for _, cell := range originalNetwork.Cells {
		copyCellToNetwork(cell, newNetwork)
	}
	for _, synapse := range originalNetwork.Synapses {
		copySynapseToNetwork(synapse, newNetwork)
	}

	if ok, report := CheckIntegrity(newNetwork); !ok {
		q.Q("CloneNetwork failed to produce network with integrity", report)
		panic("no integrity")
	}
	return newNetwork
}

/*
copyCellToNetwork copies the properies of once cell to a new one, and updates the network pointer
on the new cell to a different given network. It also adds the cell to the new network.

Returns the cell ID of the copied cell, in case it had to change.
*/
func copyCellToNetwork(cell *Cell, newNetwork *Network) CellID {
	copiedCell := NewCell(newNetwork)
	// the NewCell method automatically adds it to the network; do not allow this.
	delete(newNetwork.Cells, copiedCell.ID)

	copiedCell.ID = cell.ID

	// This is supposed to be a new cell. However, if due to an ID collision, the cell ID already
	// existed on the newNetwork, we need to change the cell ID yet again.
	if _, cellIDAlreadyOnNewNetwork := newNetwork.Cells[copiedCell.ID]; cellIDAlreadyOnNewNetwork {
		fmt.Println("warn: copyCellToNetwork would have overwritten cell with same ID")
		var newID CellID
		for {
			// new ID must be unique on both old and new network, otherwise we might get
			// incorrect wiring.
			newID = NewCellID()
			_, alreadyExistsOnNew := newNetwork.Cells[newID]
			_, alreadyExistsOnOld := cell.Network.Cells[newID]
			if !alreadyExistsOnNew && !alreadyExistsOnOld {
				break
			}
			fmt.Println("warn: copyCellToNetwork would have gotten dupe cell ID yet again")
		}
		// now change the cell ID
		copiedCell.ID = newID
	}

	newNetwork.Cells[copiedCell.ID] = copiedCell
	copiedCell.Network = newNetwork

	copiedCell.Immortal = cell.Immortal
	copiedCell.activating = cell.activating
	copiedCell.Voltage = cell.Voltage

	// golang does not copy a map on assignment; must loop over it.

	for synapseID := range cell.AxonSynapses {
		copiedCell.AxonSynapses[synapseID] = true
	}
	for synapseID := range cell.DendriteSynapses {
		copiedCell.DendriteSynapses[synapseID] = true
	}

	return copiedCell.ID
}

/*
copySynapseToNetwork copies the properies of once synapse to a new one, and updates the network pointer
on the new synapse to a different given network.

Returns its synapse ID in case it had to change.
*/
func copySynapseToNetwork(synapse *Synapse, newNetwork *Network) SynapseID {
	copiedSynapse := NewSynapse(newNetwork)
	// the NewSynapse method automatically adds it to the network; do not allow this.
	delete(newNetwork.Synapses, copiedSynapse.ID)

	copiedSynapse.ID = synapse.ID

	// This is supposed to be a new synapse. However, if due to an ID collision, the synapse ID already
	// existed on the newNetwork, we need to change the synapse ID yet again.
	if _, synapseIDAlreadyOnNewNetwork := newNetwork.Synapses[copiedSynapse.ID]; synapseIDAlreadyOnNewNetwork {
		// fmt.Println("warn: copySynapseToNetwork would have overwritten synapse with same ID")
		var newID SynapseID
		for {
			// new ID must be unique on both old and new network, otherwise we might get
			// incorrect wiring.
			newID = NewSynapseID()
			_, alreadyExistsOnNew := newNetwork.Synapses[newID]
			_, alreadyExistsOnOld := synapse.Network.Synapses[newID]
			if !alreadyExistsOnNew && !alreadyExistsOnOld {
				break
			}
			fmt.Println("warn: copySynapseToNetwork would have gotten dupe synapse ID yet again")
		}
		// now change the synapse ID
		copiedSynapse.ID = newID
	}

	newNetwork.Synapses[copiedSynapse.ID] = copiedSynapse
	copiedSynapse.Network = newNetwork

	copiedSynapse.Millivolts = synapse.Millivolts
	copiedSynapse.FromNeuronAxon = synapse.FromNeuronAxon
	copiedSynapse.ToNeuronDendrite = synapse.ToNeuronDendrite

	// synapse ID changed so we need to update its cells
	if synapse.ID != copiedSynapse.ID {
		newNetwork.Cells[copiedSynapse.ToNeuronDendrite].DendriteSynapses[copiedSynapse.ID] = true
		newNetwork.Cells[copiedSynapse.FromNeuronAxon].AxonSynapses[copiedSynapse.ID] = true
	}

	// we do need to keep this because we might want to grow the synapse later
	copiedSynapse.ActivationHistory = synapse.ActivationHistory

	return copiedSynapse.ID
}
