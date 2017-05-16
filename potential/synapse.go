package potential

import (
	"fmt"
	"math/rand"

	"github.com/ruffrey/nurtrace/laws"
)

// Synapses that fire together wire together.

/*
SynapseID should be unique for all the synapses in the network.
*/
type SynapseID uint32

/*
NewSynapseID generates a new random SynapseID
*/
func NewSynapseID() (sid SynapseID) {
	return SynapseID(rand.Uint32())
}

/*
Synapse is a construct representing the connection between two cells.

It is a one-way connection. It can either excite or inhibit the receiver.

Cell Axon -> Cell Dendrite
*/
type Synapse struct {
	ID      SynapseID
	Network *Network `json:"-"` // skip circular reference in JSON
	// Represents how many millivolts a synapse can modify the cell's
	// voltage which receives its firings.
	Millivolts        int16
	FromNeuronAxon    CellID
	ToNeuronDendrite  CellID
	ActivationHistory uint `json:"-"` // unnecessary to recreate synapse
}

/*
NewSynapse instantiates a synapse with a random millivolt weight.

It is up to the implementer to set add it to the network and set the pointer.
*/
func NewSynapse(network *Network) *Synapse {
	var id SynapseID
	for {
		id = NewSynapseID()
		if _, alreadyExists := network.Synapses[id]; !alreadyExists {
			break
		}
		// fmt.Println("warn: would have gotten dupe synapse ID")
	}
	mv := int16(randomIntBetween(laws.NewSynapseMinMillivolts, laws.NewSynapseMaxMillivolts))
	s := Synapse{
		ID:         id,
		Network:    network,
		Millivolts: mv,
	}
	synapse := &s
	network.synMux.Lock()
	network.Synapses[id] = synapse
	network.synMux.Unlock()
	return synapse
}

/*
reinforce a synapse relationship and create a new synapse of the same
direction if so.
*/
func (synapse *Synapse) reinforce() SynapseID {
	return reinforceByAmount(synapse, laws.SynapseLearnRate)
}

func reinforceByAmount(synapse *Synapse, millivolts int16) (newSynapse SynapseID) {
	mv := int16(millivolts)
	isPositive := synapse.Millivolts >= 0
	if isPositive {
		newMV := int16(synapse.Millivolts) + mv
		if newMV > laws.ActualSynapseMax {
			half := laws.ActualSynapseMax / 2
			synapse.Millivolts = half
			// add a new synapse between those two cells
			s := synapse.Network.linkCells(synapse.FromNeuronAxon, synapse.ToNeuronDendrite)
			newSynapse = s.ID
			s.Millivolts = half
		} else {
			synapse.Millivolts = newMV
		}
		return newSynapse
	}
	// negative
	newMV := int16(synapse.Millivolts) - mv
	if newMV < laws.ActualSynapseMin {
		half := laws.ActualSynapseMin / 2
		synapse.Millivolts = half
		// add a new synapse between those two cells
		s := synapse.Network.linkCells(synapse.FromNeuronAxon, synapse.ToNeuronDendrite)
		newSynapse = s.ID
		s.Millivolts = half
	} else {
		synapse.Millivolts = newMV
	}
	return newSynapse
}

func (synapse *Synapse) String() string {
	s := fmt.Sprintf("Synapse %d", synapse.ID)
	s += fmt.Sprintf("\n  Millivolts=%d", synapse.Millivolts)
	s += fmt.Sprintf("\n  ActivationHistory=%d", synapse.ActivationHistory)
	s += fmt.Sprintf("\n  FromNeuronAxon=%d", synapse.FromNeuronAxon)
	s += fmt.Sprintf("\n  ToNeuronDendrite=%d", synapse.ToNeuronDendrite)

	return s
}

/*
PruneSynapse removes a synapse and all references to it.

A synapse exists between two neurons. We get both neurons and remove the synapse
from its list.

If either of those neurons no longer has any synapses itself, kill off that neuron cell.

Unless the neuron is immortal, then just remove the synapse.
*/
func (network *Network) PruneSynapse(synapseID SynapseID) {
	// fmt.Println("remove synapse=", synapseID)
	var synapse *Synapse

	synapse, ok := network.Synapses[synapseID]
	if !ok {
		fmt.Println("warn: attempt to remove synapse that is not in network", synapseID)
		return
	}

	// See if either cell (to/from) should be pruned, also.
	// Technically this can result in a cell being the end of a dead pathway, or not receiving
	// any input. But that is something to revisit. It is likely these cells would eventually
	// build up more synapses via the grow process, or never fire and be pruned later.

	network.removeSynapseFromCell(synapseID, synapse.FromNeuronAxon, true)
	network.removeSynapseFromCell(synapseID, synapse.ToNeuronDendrite, false)

	network.synMux.Lock()
	delete(network.Synapses, synapse.ID)
	network.synMux.Unlock()
	// this synapse is now dead
}

func (network *Network) removeSynapseFromCell(s SynapseID, c CellID, isAxon bool) {
	cell, exists := network.Cells[c]
	if exists {
		network.synMux.Lock()
		if isAxon {
			delete(cell.AxonSynapses, s)
		} else {
			delete(cell.DendriteSynapses, s)
		}
		network.synMux.Unlock()
	} else {
		fmt.Println("warn: cannot prune synapse", s, "(isAxon=", isAxon, ") from cell",
			c, " cell does not exist")
		panic("referenced cell on synapse does not exist so cannot remove it")
	}

	// after removing the synapses, see if this cell can be removed.
	cellHasNoSynapses := len(cell.AxonSynapses) == 0 && len(cell.DendriteSynapses) == 0
	if cellHasNoSynapses {
		network.PruneCell(c)
	}
}
