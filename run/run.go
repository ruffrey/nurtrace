package main

import (
	"bleh/potential"
	"fmt"
)

func main() {
	network := potential.NewNetwork()
	neuronsToAdd := 25
	defaultNeuronSynapses := 10
	synapsesToAdd := 100
	network.Grow(neuronsToAdd, defaultNeuronSynapses, synapsesToAdd)
	fmt.Println("Created network")
	network.Cells[0].FireActionPotential()
	printNetwork(&network)
}

func printNetwork(network *potential.Network) {
	for _, cell := range network.Cells {
		fmt.Println("----------\ncell id=", cell.ID)
		fmt.Println("  voltage=", cell.Voltage)
		fmt.Println("  axons=", cell.AxonSynapses)
		fmt.Println("  dendrites=", cell.DendriteSynapses)
	}
}
