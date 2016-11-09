package main

// Synapses that fire together wire together.

// Represents how many millivolts a synapse can modify the cell's voltage which receives
// its firings.
const synapseMin int = -20
const synapseMax int = 20

/*
Synapse is a construct for storing how much a one-way connection between two cells will
excite or inhibit the receiver.
*/
type Synapse struct {
	Millivolts        int8
	FromNeuronAxon    *Cell
	ToNeuronDendrite  *Cell
	ActivationHistory uint
}

/*
NewSynapse instantiates a synapse with a random millivolt weight
*/
func NewSynapse() Synapse {
	return Synapse{
		Millivolts: int8(randomIntBetween(synapseMin, synapseMax)),
	}
}

/*
Activate is when the dendrite receives voltage.
*/
func (synapse *Synapse) Activate() {
	synapse.ActivationHistory++
	synapse.FromNeuronAxon.applyVoltage(synapse.Millivolts, synapse)
}
