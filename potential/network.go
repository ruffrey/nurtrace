package potential

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"
)

/*
Shasum is the sha 256 checksum of the network, used to represent its version
*/
type Shasum string

/*
Network is a full neural network
*/
type Network struct {
	Version Shasum
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
	Receptors map[int]*Receptor

	/*
	   Perceptors are sometimes known as the output of the brain. After an input is fed into the
	   receptor layer, it ripples through the Cells and a perception layer item fires.
	*/
	Perceptors map[int]*Perceptor

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
		Version:                 "0",
		Synapses:                make(map[SynapseID]*Synapse),
		Cells:                   make(map[CellID]*Cell),
		Receptors:               make(map[int]*Receptor),
		Perceptors:              make(map[int]*Perceptor),
		SynapseMinFireThreshold: 2,
		SynapseLearnRate:        1,
	}
}

/*
RegenVersion generates a new network version and sets the Version property.
*/
func (network *Network) RegenVersion() {
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%v", network)))
	network.Version = Shasum(hex.EncodeToString(hasher.Sum(nil)))
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
	// fmt.Println("Grow session start")
	// Next we move the less used synapses toward zero, because doing this later would prune the
	// brand new synapses. This is a good time to apply the learning rate to synapses
	// which were activated, too.
	var synapsesToRemove []*Synapse
	// fmt.Println("  processing learning on cells, total=", len(network.Cells))
	for _, cell := range network.Cells {
		for synapseID := range cell.DendriteSynapses { // could also be axons, but, meh.
			synapse := network.Synapses[synapseID]
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
	// fmt.Println("  done")

	// fmt.Println("  synapses to remove=", len(synapsesToRemove))
	// Actually pruning synapses is done after the previous loop because it can
	// trigger removal of Cells, which can subsequently mess up the range operation
	// happening over the same array of cells.
	for _, synapse := range synapsesToRemove {
		network.PruneSynapse(synapse)
	}
	// fmt.Println("  done")

	// fmt.Println("  adding neurons =", neuronsToAdd)
	// Now - all the new neurons are added first with no synapses. If synapses were added at
	// create time, the newer neurons would end up with far fewer connections to the following
	// newer neurons.
	var addedNeurons []*Cell
	for i := 0; i < neuronsToAdd; i++ {
		cell := NewCell(network)
		addedNeurons = append(addedNeurons, &cell)
	}
	// fmt.Println("  done")

	// fmt.Println("  adding default synapses to new neurons", defaultNeuronSynapses)
	// Now we add the default number of synapses to our new neurons, with random other neurons.
	for _, cell := range addedNeurons {
		for i := 0; i < defaultNeuronSynapses; {
			synapse := NewSynapse(network)
			ix := network.RandomCellKey()
			otherCell := network.Cells[ix]
			if cell.ID == otherCell.ID {
				// try again
				continue
			}
			if chooseIfSender() {
				synapse.FromNeuronAxon = cell.ID
				synapse.ToNeuronDendrite = otherCell.ID
				otherCell.DendriteSynapses[synapse.ID] = true
				cell.AxonSynapses[synapse.ID] = true
			} else {
				synapse.FromNeuronAxon = otherCell.ID
				synapse.ToNeuronDendrite = cell.ID
				otherCell.AxonSynapses[synapse.ID] = true
				cell.DendriteSynapses[synapse.ID] = true
			}
			// fmt.Println("created synapse", synapse)
			// must go last because it appears to copy on assignment
			network.Synapses[synapse.ID] = &synapse
			i++
		}
	}
	// fmt.Println("  done")

	// fmt.Println("  adding synapses to whole network", synapsesToAdd)
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

		synapse := NewSynapse(network)
		synapse.ToNeuronDendrite = receiver.ID
		synapse.FromNeuronAxon = sender.ID
		sender.AxonSynapses[synapse.ID] = true
		receiver.DendriteSynapses[synapse.ID] = true
		// fmt.Println("created synapse", synapse)
		// must go last because it appears to copy on assignment
		network.Synapses[synapse.ID] = &synapse
		i++
	}
	// fmt.Println("  done")

	// fmt.Println("  Grow session end")
}

/*
PruneSynapse removes a synapse and all references to it.

A synapse exists between two neurons. We get both neurons and remove the synapse
from its list.

If either of those neurons no longer has any synapses itself, kill off that neuron cell.
*/
func (network *Network) PruneSynapse(synapse *Synapse) {
	dendriteCell := network.Cells[synapse.FromNeuronAxon]
	axonCell := network.Cells[synapse.ToNeuronDendrite]

	delete(dendriteCell.AxonSynapses, synapse.ID)
	delete(axonCell.DendriteSynapses, synapse.ID)

	// See if either cell (to/from) should be pruned, also.
	// Technically this can result in a cell being the end of a dead pathway, or not receiving
	// any input. But that is something to revisit. It is likely these cells would eventually
	// build up more synapses via the grow process.
	receiverCellHasNoSynapses := len(dendriteCell.AxonSynapses) == 0 && len(dendriteCell.DendriteSynapses) == 0
	if receiverCellHasNoSynapses {
		delete(network.Cells, dendriteCell.ID)
	}
	senderCellHasNoSynapses := len(axonCell.AxonSynapses) == 0 && len(axonCell.DendriteSynapses) == 0
	if senderCellHasNoSynapses {
		delete(network.Cells, axonCell.ID)
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
RandomCellKey gets the key of a random one in the map.

This is pretty slow, as it turns out.
*/
func (network *Network) RandomCellKey() (randCellID CellID) {
	iterate := randomIntBetween(0, len(network.Cells)-1)
	i := 0
	for k := range network.Cells {
		if i == iterate {
			randCellID = CellID(k)
			break
		}
		i++
	}
	if randCellID == CellID(0) {
		panic("Should never get cell ID 0")
	}
	return randCellID
}

func chooseIfSender() bool {
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
