package potential

import (
	"github.com/ruffrey/nurtrace/laws"
	"fmt"
	"math/rand"
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
