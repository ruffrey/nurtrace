package potential

import (
	"fmt"

	"github.com/ruffrey/nurtrace/laws"
)

// Synapses that fire together wire together.

/*
SynapseID should be unique for all the synapses in the network.
*/
type SynapseID uint32

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
	fireNextRound     bool
}

/*
NewSynapse instantiates a synapse with a random millivolt weight.

It is up to the implementer to set add it to the network and set the pointer.
*/
func NewSynapse(network *Network) *Synapse {
	mv := int16(randomIntBetween(laws.NewSynapseMinMillivolts, laws.NewSynapseMaxMillivolts))

	network.synMux.Lock()
	s := Synapse{
		ID:            SynapseID(len(network.Synapses)),
		Network:       network,
		Millivolts:    mv,
		fireNextRound: false,
	}
	synapse := &s
	network.Synapses = append(network.Synapses, synapse)
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
	// log.Println("remove synapse=", synapseID)
	synapse := network.GetSyn(synapseID)

	// See if either cell (to/from) should be pruned, also.
	// Technically this can result in a cell being the end of a dead pathway, or not receiving
	// any input. But that is something to revisit. It is likely these cells would eventually
	// build up more synapses via the grow process, or never fire and be pruned later.

	network.removeSynapseFromCell(synapseID, synapse.FromNeuronAxon, true)
	network.removeSynapseFromCell(synapseID, synapse.ToNeuronDendrite, false)

	network.synMux.Lock()
	network.Synapses[synapse.ID] = nil
	network.synMux.Unlock()
	// this synapse is now dead
}

func (network *Network) removeSynapseFromCell(s SynapseID, c CellID, isAxon bool) {
	cell := network.GetCell(c)
	network.synMux.Lock()
	if isAxon {
		delete(cell.AxonSynapses, s)
	} else {
		delete(cell.DendriteSynapses, s)
	}
	network.synMux.Unlock()

	// after removing the synapses, see if this cell can be removed.
	cellHasNoSynapses := len(cell.AxonSynapses) == 0 && len(cell.DendriteSynapses) == 0
	if cellHasNoSynapses {
		network.PruneCell(c)
	}
}
