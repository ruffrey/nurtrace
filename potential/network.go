package potential

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"time"
)

/*
Network is a full neural network
*/
type Network struct {
	/*
	   Synapses are where the magic happens.
	*/
	Synapses map[SynapseID]*Synapse

	/*
		Cells are the neurons that hold the actual structure of the potential brain.
		However, with perception layers and
	*/
	Cells map[CellID]*Cell

	/*
	   Receptors are sometimes known as the input of the brain.
	*/
	Receptors map[int]Receptor

	/*
	   Perceptors are sometimes known as the output of the brain. After an input is fed into the
	   receptor layer, it ripples through the Cells and a perception layer item fires.
	*/
	Perceptors map[int]Perceptor

	// There are two factors that result in degrading a synapse:

	/*
		The minimum count of times a synapse fired in a given round between Grow cycles, with
		a millivolts=0 value.
	*/
	SynapseMinFireThreshold uint
	/*
	   How much a synapse should get bumped when it is being reinforced
	*/
	SynapseLearnRate int8
}

/*
NewNetwork is a constructor that, which also happens to reset the random number generator
when called. Seems like a good time.
*/
func NewNetwork() Network {
	rand.Seed(time.Now().Unix())
	return Network{
		Synapses:                make(map[SynapseID]*Synapse),
		Cells:                   make(map[CellID]*Cell),
		Receptors:               make(map[int]Receptor),
		Perceptors:              make(map[int]Perceptor),
		SynapseMinFireThreshold: 2,
		SynapseLearnRate:        1,
	}
}

/*
ToJSON gives a json representation of the neural network.
*/
func (network *Network) ToJSON() (string, error) {
	bytes, err := json.MarshalIndent(network, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

/*
Grow adds neurons, adds new synapses, prunes old neurons, and strengthens synapses that have
fired a lot.
*/
func (network *Network) Grow(neuronsToAdd, defaultNeuronSynapses, synapsesToAdd int) {
	// Next we move the less used synapses toward zero, because doing this later would prune the
	// brand new synapses. This is a good time to apply the learning rate to synapses
	// which were activated, too.
	var synapsesToRemove []*Synapse
	for _, cell := range network.Cells {
		for _, synapse := range cell.DendriteSynapses { // could also be axons, but, meh.
			isPositive := synapse.Millivolts >= 0

			if synapse.ActivationHistory >= network.SynapseMinFireThreshold {
				// it was activated enough, so we bump it away from zero.
				// needs cleanup refactoring.
				if isPositive {
					if int16(synapse.Millivolts)-int16(network.SynapseLearnRate) > -127 {
						synapse.Millivolts -= network.SynapseLearnRate // do not overflow int8
					} else {
						synapse.Millivolts = -128
					}
				} else {
					if int16(synapse.Millivolts)+int16(network.SynapseLearnRate) < 126 {
						synapse.Millivolts += network.SynapseLearnRate
					} else {
						synapse.Millivolts = 127
					}
				}
			} else {
				// It reached "no effect" of 0 millivolts last round. Then it didn't fire
				// this round. Remove this synapse.
				if synapse.Millivolts == 0 {
					synapsesToRemove = append(synapsesToRemove, synapse)
				}

				// did not meet minimum fire threshold, so punish it by moving toward zero
				if isPositive {
					synapse.Millivolts -= network.SynapseLearnRate
				} else {
					synapse.Millivolts += network.SynapseLearnRate
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
		network.Cells[cell.ID] = &cell
		addedNeurons = append(addedNeurons, &cell)
	}

	// Now we add the default number of synapses to our new neurons, with random other neurons.
	for _, cell := range addedNeurons {
		for i := 0; i < defaultNeuronSynapses; {
			synapse := NewSynapse()
			network.Synapses[synapse.ID] = &synapse
			ix := network.RandomCellKey()
			otherCell := network.Cells[ix]
			if cell.ID == otherCell.ID {
				// try again
				continue
			}
			if makeSender() {
				synapse.FromNeuronAxon = cell
				synapse.ToNeuronDendrite = otherCell
				otherCell.DendriteSynapses = append(otherCell.DendriteSynapses, &synapse)
				cell.AxonSynapses = append(cell.AxonSynapses, &synapse)
			} else {
				synapse.FromNeuronAxon = otherCell
				synapse.ToNeuronDendrite = cell
				otherCell.AxonSynapses = append(otherCell.AxonSynapses, &synapse)
				cell.DendriteSynapses = append(cell.DendriteSynapses, &synapse)
			}
			i++
		}
	}

	// Then we randomly add synapses between neurons to the whole network, including the
	// newest neurons.
	for i := 0; i < synapsesToAdd; {
		senderIx := network.RandomCellKey()
		receiverIx := network.RandomCellKey()
		sender := network.Cells[senderIx]
		receiver := network.Cells[receiverIx]
		// Thy cell shannot activate thyself
		if sender.ID == receiver.ID {
			continue
		}

		synapse := NewSynapse()
		network.Synapses[synapse.ID] = &synapse
		synapse.ToNeuronDendrite = receiver
		synapse.FromNeuronAxon = sender
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
		if reflect.DeepEqual(ref, synapse.FromNeuronAxon) {
			synapse.FromNeuronAxon.AxonSynapses = append(synapse.FromNeuronAxon.AxonSynapses[:index], synapse.FromNeuronAxon.AxonSynapses[index+1:]...)
			removedAxonRef = true
			break
		}
	}
	if !removedAxonRef {
		panic("Failed to remove a synapse refernece from the AXON side of a cell")
	}

	for index, ref := range synapse.ToNeuronDendrite.DendriteSynapses {
		if reflect.DeepEqual(ref, synapse.ToNeuronDendrite) {
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
		for key, cell := range network.Cells {
			if cell.ID == synapse.ToNeuronDendrite.ID {
				network.PruneNeuron(key)
				break
			}
		}
	}
	if len(synapse.FromNeuronAxon.AxonSynapses) == 0 && len(synapse.FromNeuronAxon.DendriteSynapses) == 0 {
		// find it's key and remove it right now
		for key, cell := range network.Cells {
			if cell.ID == synapse.FromNeuronAxon.ID {
				network.PruneNeuron(key)
				break
			}
		}
	}

	delete(network.Synapses, synapse.ID)
	// this synapse is now dead
}

/*
PruneNeuron deactivates and removes a neuron cell.

It is assumed this neuron has no synapses!

It also cannot be called in a range operation over network.Cells, because it will be removing
the cell at the supplied index.
*/
func (network *Network) PruneNeuron(key CellID) {
	cell := network.Cells[key]
	if len(cell.DendriteSynapses) != 0 || len(cell.AxonSynapses) != 0 {
		panic("Attempting to prune a neuron which still has synapses")
	}
	delete(network.Cells, key)
}

func randomIntBetween(min, max int) int {
	return rand.Intn((max+1)-min) + min
}

/*
Methods for random map keys below select a random integer between 0 and the map length,
then interate into the map that many times to find the map key we want.
Golang technically will not guarantee looping order over a map; but
it is not truly random, so we have to do this expensive task of looping.
It could likely be improved later by using a larger memory footprint
that tracks all the keys in an array of integers.
*/

/*
RandomCellKey gets the key of a random one in the map
*/
func (network *Network) RandomCellKey() CellID {
	iterate := randomIntBetween(0, len(network.Cells)-1)
	i := 0
	for k := range network.Cells {
		if i >= iterate {
			return CellID(k)
		}
		i++
	}
	return CellID(0)
}

func makeSender() bool {
	if randomIntBetween(0, 1) == 1 {
		return true
	}
	return false
}

/*
Equilibrium should run periodically to bring all cells closer to their apResting voltage.
*/
func (network *Network) Equilibrium() {
	for _, cell := range network.Cells {
		if cell.Voltage != apResting {
			diff := cell.Voltage - apResting
			// get halfway to resting
			cell.Voltage -= (diff / 2)
		}
	}
}

/*
PrintCells logs the network cells to console
*/
func (network *Network) PrintCells() {
	for _, cell := range network.Cells {
		fmt.Println("----------\ncell id=", cell.ID)
		fmt.Println("  voltage=", cell.Voltage)
		fmt.Println("  axons=", cell.AxonSynapses)
		fmt.Println("  dendrites=", cell.DendriteSynapses)
	}
}

/*
SaveToFile outputs the network to a file
*/
func (network *Network) SaveToFile(filepath string) (err error) {
	contents, err := network.ToJSON()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath, []byte(contents), os.ModePerm)
	return nil
}
