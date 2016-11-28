package main

import (
	"bleh/potential"
	"fmt"
	"time"
)

func main() {
	network := potential.NewNetwork()
	neuronsToAdd := 50
	defaultNeuronSynapses := 10
	synapsesToAdd := 100
	network.Grow(neuronsToAdd, defaultNeuronSynapses, synapsesToAdd)
	network.RegenVersion()
	fmt.Println("Created network Version=", network.Version)
	printNetwork(&network)
	iterations := 1000000
	for i := 1; i < iterations; i++ {
		if i%10000 == 0 {
			var p float64
			p = float64(i) / float64(iterations)
			fmt.Println(" progress", p)
		}
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
