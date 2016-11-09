package main

import (
	"math/rand"
	"reflect"
	"time"
)

// There are two factors that result in battering down a synapse:

// The minimum count of times a synapse fired in a given round between Grow cycles, with
// a millivolts=0 value.
const synapseMinFireThreshold uint = 2

// How much a synapse should get bumped
const synapseLearnRate int8 = 1

/*
Network is a neural network
*/
type Network struct {
	Cells []Cell
}

/*
NewNetwork is a constructor that happens to reseet the random number generator.
*/
func NewNetwork() Network {
	rand.Seed(time.Now().Unix())
	return Network{}
}

/*
Grow adds neurons, adds new synapses, prunes old neurons, and strengthens synapses that have
fired a lot.
*/
func (network *Network) Grow(neuronsToAdd, defaultNeuronSynapses, synapsesToAdd, minSynapseFiringKeepCount int) {
	// Next we move the less used synapses toward zero, because doing this later would prune the
	// brand new synapses. This is a good time to apply the learning rate to synapses
	// which were activated, too.
	var synapsesToRemove []*Synapse
	for _, cell := range network.Cells {
		for _, synapse := range cell.DendriteSynapses { // could also be axons, but, meh.
			isPositive := synapse.Millivolts >= 0

			if synapse.ActivationHistory >= synapseMinFireThreshold {
				// it was activated enough, so we bump it away from zero.
				if isPositive {
					synapse.Millivolts -= synapseLearnRate
				} else {
					synapse.Millivolts += synapseLearnRate
				}
			} else {
				// It reached "no effect" of 0 millivolts last round. Then it didn't fire
				// this round. Remove this synapse.
				if synapse.Millivolts == 0 {
					synapsesToRemove = append(synapsesToRemove, synapse)
				}

				// did not meet minimum fire threshold, so punish it by moving toward zero
				if isPositive {
					synapse.Millivolts -= synapseLearnRate
				} else {
					synapse.Millivolts += synapseLearnRate
				}

				// Next time, if it did not fire, and it is zero, it will get pruned.
			}

			// reset the activation history until the next Grow cycle.
			synapse.ActivationHistory = 0
		}
	}
	// Actually pruning synapses is done after the previous loop because it can
	// trigger removal of Cells, which can subsequently mess up the range operation
	// happening over the same array of cells.
	for _, synapse := range synapsesToRemove {
		network.PruneSynapse(synapse)
	}

	// Now - all the new neurons are added first with no synapses. If synapses were added at
	// create time, the newer neurons would end up with far fewer connections to the following
	// newer neurons.
	var addedNeurons []*Cell
	for i := 0; i < neuronsToAdd; i++ {
		cell := NewCell()
		network.Cells = append(network.Cells, cell)
		addedNeurons = append(addedNeurons, &cell)
	}

	// Now we add the default number of synapses to our new neurons, with random other neurons.
	for _, cell := range addedNeurons {
		synapse := NewSynapse()
		otherCell := network.Cells[randomIntBetween(0, len(network.Cells))]
		if is(cell, otherCell) {
			// TODO: decide what do to here, when attempting to add a synapse to a new neuron,
			// but the randomIntBetween call picked the same neuron.
			continue
		}
		if makeSender() {
			synapse.FromNeuronAxon = cell
			synapse.ToNeuronDendrite = &otherCell
			otherCell.DendriteSynapses = append(otherCell.DendriteSynapses, &synapse)
			cell.AxonSynapses = append(cell.AxonSynapses, &synapse)
		} else {
			synapse.FromNeuronAxon = &otherCell
			synapse.ToNeuronDendrite = cell
			otherCell.AxonSynapses = append(otherCell.AxonSynapses, &synapse)
			cell.DendriteSynapses = append(cell.DendriteSynapses, &synapse)
		}
	}

	// Then we randomly add synapses between neurons to the whole network, including the
	// newest neurons.
	for i := 0; i < synapsesToAdd; {
		sender := network.Cells[randomIntBetween(0, len(network.Cells))]
		receiver := network.Cells[randomIntBetween(0, len(network.Cells))]
		// Thy cell shannot activate thyself
		if is(sender, receiver) {
			continue
		}

		synapse := NewSynapse()
		synapse.ToNeuronDendrite = &receiver
		synapse.FromNeuronAxon = &sender
		sender.AxonSynapses = append(sender.AxonSynapses, &synapse)
		receiver.DendriteSynapses = append(receiver.DendriteSynapses, &synapse)
		i++
	}

}

/*
PruneSynapse removes a synapse and all references to it. A synapse only exists by being
referenced by neurons. So we remove those references and hope the garbage collector
cleans it up.

If either of those neurons no longer has any synapses itself, kill off that neuron cell.
*/
func (network *Network) PruneSynapse(synapse *Synapse) {
	removedAxonRef := false
	removedDendriteRef := false

	for index, ref := range synapse.FromNeuronAxon.AxonSynapses {
		if is(ref, synapse.FromNeuronAxon) {
			synapse.FromNeuronAxon.AxonSynapses = append(synapse.FromNeuronAxon.AxonSynapses[:index], synapse.FromNeuronAxon.AxonSynapses[index+1:]...)
			removedAxonRef = true
			break
		}
	}
	if !removedAxonRef {
		panic("Failed to remove a synapse refernece from the AXON side of a cell")
	}

	for index, ref := range synapse.ToNeuronDendrite.DendriteSynapses {
		if is(ref, synapse.ToNeuronDendrite) {
			synapse.ToNeuronDendrite.DendriteSynapses = append(synapse.ToNeuronDendrite.DendriteSynapses[:index], synapse.ToNeuronDendrite.DendriteSynapses[index+1:]...)
			removedDendriteRef = true
			break
		}
	}
	if !removedDendriteRef {
		panic("Failed to remove a synapse refernece from the DENDTIRE side of a cell")
	}

	// See if either cell (to/from) should be pruned, also.
	// Technically this can result in a cell being the end of a dead pathway, or not receiving
	// any input. But that is something to revisit. It is likely these cells would eventually
	// build up more synapses via the grow process.
	if len(synapse.ToNeuronDendrite.AxonSynapses) == 0 && len(synapse.ToNeuronDendrite.DendriteSynapses) == 0 {
		// find it's index and remove it right now
		for index, cell := range network.Cells {
			if is(cell, synapse.ToNeuronDendrite) {
				network.PruneNeuron(index)
				break
			}
		}
	}
	if len(synapse.FromNeuronAxon.AxonSynapses) == 0 && len(synapse.FromNeuronAxon.DendriteSynapses) == 0 {
		// find it's index and remove it right now
		for index, cell := range network.Cells {
			if is(cell, synapse.FromNeuronAxon) {
				network.PruneNeuron(index)
				break
			}
		}
	}

	// this synapse is now dead
}

/*
PruneNeuron deactivates and removes a neuron cell.

It is assumed this neuron has no synapses!

It also cannot be called in a range operation over network.Cells, because it will be removing
the cell at the supplied index.
*/
func (network *Network) PruneNeuron(index int) {
	cell := network.Cells[index]
	if len(cell.DendriteSynapses) != 0 || len(cell.AxonSynapses) != 0 {
		panic("Attempting to prune a neuron which still has synapses")
	}
	network.Cells = append(network.Cells[:index], network.Cells[index+1:]...) // this removes it
	cell.Destroy()
}

func randomIntBetween(min, max int) int {
	return rand.Intn(max-min) + min
}

func makeSender() bool {
	if randomIntBetween(0, 1) == 1 {
		return true
	}
	return false
}

func is(a, b interface{}) bool {
	return reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer()
}
