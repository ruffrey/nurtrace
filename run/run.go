package main

import (
	"bleh/potential"
	"fmt"
	"time"
)

func main() {
	network := potential.NewNetwork()
	neuronsToAdd := 5000
	defaultNeuronSynapses := 10
	synapsesToAdd := 100
	network.Grow(neuronsToAdd, defaultNeuronSynapses, synapsesToAdd)
	network.RegenVersion()
	fmt.Println("Created network Version=", network.Version)
	printNetwork(&network)
	for i := 0; i < 10000000; i++ {
		cellID := network.RandomCellKey()
		cell := network.Cells[cellID]
		cell.FireActionPotential()
	}
	fmt.Println("\nAfter activation\n ")
	printNetwork(&network)
	network.Equilibrium()
	time.Sleep(2 * time.Second)
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
