package main

import (
	"bleh/potential"
	"fmt"
)

func main() {
	network := potential.NewNetwork()
	neuronsToAdd := 4
	defaultNeuronSynapses := 4
	synapsesToAdd := 10
	network.Grow(neuronsToAdd, defaultNeuronSynapses, synapsesToAdd)
	fmt.Println(network.ToJSON())
}
