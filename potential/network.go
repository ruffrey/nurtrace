package potential

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
)

/*
Network is a full neural network
*/
type Network struct {
	/*
		Indicates whether neurons are allowed to fire. Setting it disabled will stop firing
		in a few milliseconds.
	*/
	Disabled bool
	/*
	   Synapses are where the magic happens.
	*/
	Synapses map[SynapseID]*Synapse
	synMux   sync.Mutex
	/*
		Cells are the neurons that hold the actual structure of the potential brain.
		However, with perception layers and
	*/
	Cells map[CellID]*Cell

	// private

	// nextSynapsesToActivate will fire their axon cell on the next step. always true
	nextSynapsesToActivate map[SynapseID]bool

	resetCellsOnNextStep map[CellID]bool
}

/*
NewNetwork is a constructor that, which also happens to reset the random number generator
when called. Seems like a good time.
*/
func NewNetwork() Network {
	return Network{
		Disabled: false,
		Synapses: make(map[SynapseID]*Synapse),
		Cells:    make(map[CellID]*Cell),
		nextSynapsesToActivate: make(map[SynapseID]bool),
		resetCellsOnNextStep:   make(map[CellID]bool),
	}
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
	lenCells := len(network.Cells)
	if lenCells == 0 {
		panic("Cannot call RandomCellKey() on network with no cells")
	}
	iterate := randomIntBetween(0, lenCells-1)
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

/*
ResetForTraining resets transient properties on the network to their base resting state.

Does NOT reset synapse activation history - those should be reset only during a network.Prune()
cycle.
*/
func (network *Network) ResetForTraining() {
	for _, cell := range network.Cells {
		cell.activating = false
		cell.WasFired = false
		cell.Voltage = apResting
	}
	network.nextSynapsesToActivate = make(map[SynapseID]bool)
	network.resetCellsOnNextStep = make(map[CellID]bool)
	network.Disabled = false
}

/*
Print logs the network cells to console
*/
func (network *Network) Print() {
	fmt.Println("----------")
	fmt.Println("Network")
	for id, cell := range network.Cells {
		fmt.Println("  --------\n  cell key=", id, "ID=", cell.ID)
		fmt.Println("  voltage=", cell.Voltage)
		fmt.Println("  synapses to axon=", cell.AxonSynapses)
		fmt.Println("  synapses to dendrite=", cell.DendriteSynapses)
	}
	for id, syn := range network.Synapses {
		fmt.Println("  --------\n  synapse key=", id, "ID=", syn.ID)
		fmt.Println("  millivolts=", syn.Millivolts)
		fmt.Println("  axon=", syn.FromNeuronAxon)
		fmt.Println("  dendrite=", syn.ToNeuronDendrite)
	}
	fmt.Println("----------")
}

/*
PrintTotals lists the most basic info about the network.
*/
func (network *Network) PrintTotals() {
	fmt.Println("Network")
	fmt.Println(" ", len(network.Cells), "cells")
	fmt.Println(" ", len(network.Synapses), "synapses")
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
SaveToFile outputs the network to a file
*/
func (network *Network) SaveToFile(filepath string) (err error) {
	contents, err := network.ToJSON()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath, []byte(contents), os.ModePerm)
	return err
}

/*
LoadNetworkFromFile reads a saved network from disk and creates a new network from it.
*/
func LoadNetworkFromFile(filepath string) (*Network, error) {
	n := NewNetwork()
	network := &n
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return network, err
	}
	err = json.Unmarshal(bytes, network)
	if err != nil {
		return network, err
	}
	for _, synapse := range network.Synapses {
		synapse.Network = network
	}
	for _, cell := range network.Cells {
		cell.Network = network
	}

	if ok, report := CheckIntegrity(network); !ok {
		report.Print()
		return network, fmt.Errorf("Failed loading network from file %s", filepath)
	}
	return network, nil
}
