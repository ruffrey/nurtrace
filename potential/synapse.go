package potential

import "math/rand"

// Synapses that fire together wire together.

// Represents how many millivolts a synapse can modify the cell's voltage which receives
// its firings.
const synapseMin int = -10
const synapseMax int = 10

/*
Synapse is a construct for storing how much a one-way connection between two cells will
excite or inhibit the receiver.

Cell Axon -> Cell Dendrite
*/
type Synapse struct {
	ID                int
	Millivolts        int8
	FromNeuronAxon    *Cell
	ToNeuronDendrite  *Cell
	ActivationHistory uint
}

/*
NewSynapse instantiates a synapse with a random millivolt weight
*/
func NewSynapse() Synapse {
	mv := int8(randomIntBetween(synapseMin, synapseMax))
	return Synapse{
		ID:         rand.Int(),
		Millivolts: mv,
	}
}

/*
Activate is when the dendrite receives voltage.
*/
func (synapse *Synapse) Activate() {
	synapse.ActivationHistory++
	synapse.ToNeuronDendrite.ApplyVoltage(synapse.Millivolts, synapse)
}
