package main

import (
	"bleh/charrnn"
	"bleh/potential"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

func main() {
	threads := 10

	var network potential.Network
	var err error
	network, err = potential.LoadNetworkFromFile("network.json")
	if err != nil {
		fmt.Println("No existing network in file; creating new one.", err)
		network = potential.NewNetwork()
		neuronsToAdd := 1000
		defaultNeuronSynapses := 20
		synapsesToAdd := 0
		network.Grow(neuronsToAdd, defaultNeuronSynapses, synapsesToAdd)
		fmt.Println("Created network")
		fmt.Println("Saving to disk")
		err = network.SaveToFile("network.json")
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		fmt.Println("Loaded network from disk,", len(network.Cells), "cells")
	}

	bytes, err := ioutil.ReadFile("shake.txt")
	if err != nil {
		panic(err)
	}

	text := string(bytes)
	lines := strings.Split(text, "\n")

	vocab := charrnn.NewVocab(text, &network)

	fmt.Println("Loaded vocab, lenth=", len(vocab))

	fmt.Println("Beginning training", threads, "simultaneous sessions")

	totalLines := len(lines)
	for i := 0; i < totalLines; {

		for thread := 0; thread < threads; thread++ {
			if i >= totalLines { // ran out of lines
				break
			}
			line := lines[i]
			i++

			ch := make(chan bool)
			copiedNetwork := potential.CloneNetwork(&network)
			fmt.Println("starting thread", thread)
			go func() {
				processLine(line, &copiedNetwork, vocab, ch)
			}()
			success := <-ch
			if success { // keep the training
				diff := potential.DiffNetworks(&network, &copiedNetwork)
				potential.ApplyDiff(diff, &network)
				// network.Grow(0, 0, 0) // prune
			} else {
				// We failed to generate the desired effect, so do a significant growth
				// of cells.
				// network.Grow(100, 20, 200)
			}
		}
	}

	err = network.SaveToFile("network.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Failed saving network")
}

/*
processLine fires this entire line in the neural network at once, hoping to get the desired output.

It will not add any synapses.
*/
func processLine(line string, network *potential.Network, vocab charrnn.Vocab, ch chan bool) {
	lineChars := strings.Split(line, "")

	succeeded := 0

	// First time through, fire the receptors a bunch to stimulate the network,
	// and see if it resulted in the expected outputs firing.
	for ix, char := range lineChars {
		isFirst := ix == 0
		isLast := ix == (len(lineChars) - 1)
		var inChar charrnn.VocabItem
		var outChar charrnn.VocabItem
		if isFirst {
			inChar = vocab["START"]
			outChar = vocab[char]
		} else if isLast {
			inChar = vocab[char]
			outChar = vocab["END"]
		} else {
			inChar = vocab[lineChars[ix-1]]
			outChar = vocab[char]
		}

		network.ResetForTraining()
		// just make the network fires several times in order
		for i := 0; i < 10; i++ {
			doneChan := make(chan bool)
			go func() {
				time.Sleep(potential.RefractoryPeriodMillis * time.Millisecond)
				network.Cells[inChar.OutputCell].FireActionPotential()
				doneChan <- true
			}()
			<-doneChan
		}

		if network.Cells[outChar.InputCell].WasFired {
			succeeded++
		}
	}
	wasSuccessful := succeeded == (len(lineChars) - 1)
	fmt.Println(wasSuccessful, succeeded, "/", (len(lineChars) - 1), "\n  ", line)
	ch <- wasSuccessful
}
