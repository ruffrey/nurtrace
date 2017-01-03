package potential

import (
	"fmt"
	"math/rand"
)

// Synapses that fire together wire together.

// Represents how many millivolts a synapse can modify the cell's voltage which receives
// its firings.
const synapseMin int = -15
const synapseMax int = 15

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
	ID                SynapseID
	Network           *Network `json:"-"` // skip circular reference in JSON
	Millivolts        int8
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
	mv := int8(randomIntBetween(synapseMin, synapseMax))
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
Activate indicates an axon has fired and this synapse should pass the message along to
its dendrite cell.
*/
func (synapse *Synapse) Activate() (didFire bool, err error) {
	didFire = false
	synapse.ActivationHistory++
	dendriteCell, exists := synapse.Network.Cells[synapse.ToNeuronDendrite]
	if !exists {
		err = fmt.Errorf("bad synapse dendrite: not on network; dendrite=%d", synapse.ToNeuronDendrite)
		return didFire, err
	}
	// fmt.Println("  synapse", synapse.ID, "firing dendrite on", synapse.ToNeuronDendrite)
	didFire = dendriteCell.ApplyVoltage(synapse.Millivolts, synapse)

	return didFire, nil
}

/*
reinforce a synapse relationship and create a new synapse of the same
direction if so.
*/
func (synapse *Synapse) reinforce() SynapseID {
	return reinforceByAmount(synapse, synapseLearnRate)
}

func reinforceByAmount(synapse *Synapse, mv int8) (newSynapse SynapseID) {
	isPositive := synapse.Millivolts >= 0
	if isPositive {
		newMV := synapse.Millivolts + mv
		if newMV > actualSynapseMax {
			half := actualSynapseMax / 2
			synapse.Millivolts = half
			// add a new synapse between those two cells
			s := NewSynapse(synapse.Network)
			newSynapse = s.ID
			s.Millivolts = half
			synapse.Network.Cells[synapse.ToNeuronDendrite].addDendrite(newSynapse)
			synapse.Network.Cells[synapse.FromNeuronAxon].addAxon(newSynapse)
		} else {
			synapse.Millivolts = newMV
		}
		return newSynapse
	}
	// negative
	newMV := synapse.Millivolts - mv
	if newMV < actualSynapseMin {
		half := actualSynapseMin / 2
		synapse.Millivolts = half
		// add a new synapse between those two cells
		s := NewSynapse(synapse.Network)
		newSynapse = s.ID
		s.Millivolts = half
		synapse.Network.Cells[synapse.ToNeuronDendrite].addDendrite(newSynapse)
		synapse.Network.Cells[synapse.FromNeuronAxon].addAxon(newSynapse)
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
