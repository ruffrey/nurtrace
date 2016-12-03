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
		Indicates whether neurons are allowed to fire. Setting it disabled will stop firing
		in a few milliseconds.
	*/
	Disabled bool
	/*
	   Synapses are where the magic happens.
	*/
	Synapses map[SynapseID]*Synapse

	/*
		Cells are the neurons that hold the actual structure of the potential brain.
		However, with perception layers and
	*/
	Cells map[CellID]*Cell

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

	// private

	// actualSynapseMin and Max helps make math less intensive if there is never a chance synapse
	// addition will create an int8 overflow.
	actualSynapseMin int8
	actualSynapseMax int8
}

/*
NewNetwork is a constructor that, which also happens to reset the random number generator
when called. Seems like a good time.
*/
func NewNetwork() Network {
	rand.Seed(time.Now().Unix())
	lr := int8(1)
	return Network{
		Version:  "0",
		Disabled: false,
		Synapses: make(map[SynapseID]*Synapse),
		Cells:    make(map[CellID]*Cell),
		SynapseMinFireThreshold: 4,
		SynapseLearnRate:        lr,
		actualSynapseMin:        int8(-128) + lr,
		actualSynapseMax:        int8(127) - lr,
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

/*
ResetForTraining resets transient properties on the network to their base resting state.
*/
func (network *Network) ResetForTraining() {
	for _, cell := range network.Cells {
		cell.activating = false
		cell.WasFired = false
		cell.Voltage = apResting
	}
	network.Disabled = false
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
ToJSON gives a json representation of the neural network.
*/
func (network *Network) ToJSON() (string, error) {
	bytes, err := json.Marshal(network)
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
	return network, nil
}
