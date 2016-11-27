package main

import (
	"bleh/potential"
	"fmt"
)

func main() {
	var network potential.Network
	var err error
	network, err = potential.LoadNetworkFromFile("network.json")
	if err != nil {
		fmt.Println("No existing network in file; creating new one.", err)
		network = potential.NewNetwork()
		neuronsToAdd := 10
		defaultNeuronSynapses := 5
		synapsesToAdd := 30
		network.Grow(neuronsToAdd, defaultNeuronSynapses, synapsesToAdd)
		fmt.Println("Created network")
		network.PrintCells()
		fmt.Println("Saving to disk")
		err = network.SaveToFile("network.json")
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	network.PrintCells()

}
