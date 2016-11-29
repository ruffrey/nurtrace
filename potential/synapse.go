package potential

import (
	"fmt"
	"math/rand"
)

// Synapses that fire together wire together.

// Represents how many millivolts a synapse can modify the cell's voltage which receives
// its firings.
const synapseMin int = -10
const synapseMax int = 10

/*
SynapseID is just a normal Go integer (probably int64).
*/
type SynapseID int

/*
NewSynapseID generates a new random SynapseID
*/
func NewSynapseID() (sid SynapseID) {
	return SynapseID(rand.Int())
}

/*
Synapse is a construct for storing how much a one-way connection between two cells will
excite or inhibit the receiver.

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
func NewSynapse() *Synapse {
	mv := int8(randomIntBetween(synapseMin, synapseMax))
	s := Synapse{
		ID:         NewSynapseID(),
		Millivolts: mv,
	}
	synapse := &s
	return synapse
}

/*
Activate is when the dendrite receives voltage.
*/
func (synapse *Synapse) Activate() {
	synapse.ActivationHistory++
	dendriteCell, exists := synapse.Network.Cells[synapse.ToNeuronDendrite]
	if !exists {
		fmt.Println("bad synapse: ", synapse)
		panic("Attempting to activate non-existent cell connected to synapse.")
	}
	// fmt.Println("  synapse", synapse.ID, "firing dendrite on", synapse.ToNeuronDendrite)
	dendriteCell.ApplyVoltage(synapse.Millivolts, synapse)
}
