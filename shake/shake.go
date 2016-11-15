package main

import (
	"bleh/potential"
	"fmt"
)

func main() {
	network := potential.NewNetwork()
	neuronsToAdd := 50
	defaultNeuronSynapses := 10
	synapsesToAdd := 100
	network.Grow(neuronsToAdd, defaultNeuronSynapses, synapsesToAdd)
	fmt.Println("Created network")

}
