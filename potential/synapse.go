package potential

import "math/rand"

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
	Network           *Network
	Millivolts        int8
	FromNeuronAxon    CellID
	ToNeuronDendrite  CellID
	ActivationHistory uint
}

/*
NewSynapse instantiates a synapse with a random millivolt weight
*/
func NewSynapse(network *Network) Synapse {
	mv := int8(randomIntBetween(synapseMin, synapseMax))
	return Synapse{
		ID:         NewSynapseID(),
		Network:    network,
		Millivolts: mv,
	}
}

/*
Activate is when the dendrite receives voltage.
*/
func (synapse *Synapse) Activate() {
	synapse.ActivationHistory++
	dendriteCell := synapse.Network.Cells[synapse.ToNeuronDendrite]
	dendriteCell.ApplyVoltage(synapse.Millivolts, synapse)
}
