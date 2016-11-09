package main

import (
	"bleh/potential"
	"fmt"
	"time"
)

func main() {
	network := potential.NewNetwork()
	neuronsToAdd := 5
	defaultNeuronSynapses := 10
	synapsesToAdd := 10
	network.Grow(neuronsToAdd, defaultNeuronSynapses, synapsesToAdd)
	fmt.Println("Created network")
	// printNetwork(&network)
	network.Cells[0].FireActionPotential()
	network.Cells[0].FireActionPotential()
	network.Cells[0].FireActionPotential()
	network.Cells[0].FireActionPotential()
	network.Cells[0].FireActionPotential()
	fmt.Println("\nAfter activation\n ")
	printNetwork(&network)
	network.Equilibrium()
	time.Sleep(500 * time.Millisecond)
	network.Equilibrium()
	fmt.Println("\nFinal\n ")
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
