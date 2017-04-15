package potential

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"sync"

	"github.com/ruffrey/nurtrace/laws"
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
	Cells   map[CellID]*Cell
	cellMux sync.Mutex

	// private

	// nextSynapsesToActivate will fire their axon cell on the next step. always true
	nextSynapsesToActivate map[SynapseID]bool

	resetCellsOnNextStep map[CellID]bool
}

/*
NewNetwork is a constructor that, which also happens to reset the random number generator
when called. Seems like a good time.
*/
func NewNetwork() *Network {
	n := Network{
		Disabled: false,
		Synapses: make(map[SynapseID]*Synapse),
		Cells:    make(map[CellID]*Cell),
		nextSynapsesToActivate: make(map[SynapseID]bool),
		resetCellsOnNextStep:   make(map[CellID]bool),
	}
	return &n
}

/*
linkCells creates a new synapse and links the two referenced cells where the
"to" cell has an axon firing the "from" cell's dendrite.
*/
func (network *Network) linkCells(fromCellID CellID, toCellID CellID) *Synapse {
	fromCell := network.getCell(fromCellID)
	toCell := network.getCell(toCellID)

	synapse := NewSynapse(network)
	fromCell.addAxon(synapse.ID)
	toCell.addDendrite(synapse.ID)

	return synapse
}

/*
getCell safely returns a cell object so you don't have to use mutexes.
*/
func (network *Network) getCell(cellID CellID) *Cell {
	network.cellMux.Lock()
	cell := network.Cells[cellID]
	network.cellMux.Unlock()
	return cell
}

/*
getSyn safely returns a synapse object so you don't have to use mutexes.
*/
func (network *Network) getSyn(synapseID SynapseID) *Synapse {
	network.synMux.Lock()
	synapse := network.Synapses[synapseID]
	network.synMux.Unlock()
	return synapse
}

func randomIntBetween(min, max int) int {
	return rand.Intn((max+1)-min) + min
}

// randCell returns a random CellID from a map where cells are the keys.
// probably could combine with RandCellKey
func randCellFromMap(cellMap map[CellID]bool) (randCellID CellID) {
	iterate := randomIntBetween(0, len(cellMap)-1)
	i := 0
	for k := range cellMap {
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

/*
ResetForTraining resets transient properties on the network to their base resting state.

Does NOT reset synapse activation history - those should be reset only during a network.Prune()
cycle. Additionally DOES NOT reset voltage, as this is part of the network's memory at work.
*/
func (network *Network) ResetForTraining() {
	for _, cell := range network.Cells {
		cell.activating = false
		cell.WasFired = false
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
	lenCells := len(network.Cells)
	lenSynapses := len(network.Synapses)
	fmt.Println("Network")
	fmt.Println(" ", lenCells, "cells")
	fmt.Println(" ", lenSynapses, "synapses")
	fmt.Println(" ", lenSynapses/lenCells, "avg synapses per cell")
}

/*
ToJSON gives a json representation of the neural network.
*/
func (network *Network) ToJSON() ([]byte, error) {
	bytes, err := json.Marshal(network)
	if err != nil {
		return []byte{}, err
	}
	return bytes, nil
}

/*
SaveToFile outputs the network to a file as gzipped JSON
*/
func (network *Network) SaveToFile(filepath string) (err error) {
	contents, err := network.ToJSON()
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	gzipper, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return err
	}
	gzipper.Name = filepath
	gzipper.Comment = "nurtrace gz JSON"
	_, err = gzipper.Write(contents)
	if err != nil {
		return err
	}
	err = gzipper.Close()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath, buf.Bytes(), os.ModePerm)
	return err
}

/*
SaveToFileReadable outputs the network to a file as gzipped JSON
*/
func (network *Network) SaveToFileReadable(filepath string) (err error) {
	bytes, err := json.MarshalIndent(network, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath, bytes, os.ModePerm)
	return err
}

/*
FireNoise chooses `NoiseRatio` random cells and fires them.
*/
func (network *Network) FireNoise() {
	totalFires := int(math.Ceil(float64(len(network.Cells)) * laws.NoiseRatio))
	for i := 0; i < totalFires; i++ {
		network.getCell(network.RandomCellKey()).FireActionPotential()
	}
}

/*
LoadNetworkFromFile reads a saved network from disk and creates a new network from it.
*/
func LoadNetworkFromFile(filepath string) (*Network, error) {
	network := NewNetwork()
	file, err := os.Open(filepath)
	if err != nil {
		return network, err
	}
	defer file.Close()

	// try to unzip it, but it's ok if that fails, it might be regular json
	var gunzipppedBytes []byte
	gzReader, err := gzip.NewReader(file)
	if err == nil {
		defer gzReader.Close()
		gunzipppedBytes, err = ioutil.ReadAll(gzReader)
		if err != nil {
			return network, err
		}
	}

	var jsonBytes []byte
	if len(gunzipppedBytes) > 0 {
		jsonBytes = gunzipppedBytes
	} else {
		err = file.Close()
		if err != nil {
			return network, err
		}
		jsonBytes, err = ioutil.ReadFile(filepath)
		if err != nil {
			return network, err
		}
	}

	err = json.Unmarshal(jsonBytes, network)
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
		return network, fmt.Errorf("Cannot load network with bad integrity from file %s", filepath)
	}
	return network, nil
}
