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
		threadsFinished := 0

		// train in parallel over this number of threads
		ch := make(chan processResult)
		networkCopies := make(map[int]*potential.Network)
		results := make(map[int]processResult)
		for thread := 0; thread < threads; thread++ {
			if i >= totalLines { // ran out of lines
				threadsFinished = threads
				break
			}
			line := lines[i]
			copiedNetwork := potential.CloneNetwork(&network)
			networkCopies[i] = &copiedNetwork
			i++
			fmt.Println("starting thread", thread)
			go func() {
				processLine(thread, line, &copiedNetwork, vocab, ch)
			}()

		}

		// read results from the threads as they come in
		for result := range ch {
			threadsFinished++
			results[result.threadIndex] = result
			if threadsFinished >= threads {
				close(ch)
			}
		}

		// now that all threads are finished, read their results and modify the network in
		// series

		// let the network finish firing
		time.AfterFunc(100*time.Millisecond, func() {
			for thread := 0; thread < threads; thread++ {
				r := results[thread]
				if r.succeeded { // keep the training
					copiedNetwork := networkCopies[r.threadIndex]
					diff := potential.DiffNetworks(&network, copiedNetwork)
					potential.ApplyDiff(diff, &network)
					network.Grow(0, 0, 0) // prune
				} else {
					// We failed to generate the desired effect, so do a significant growth
					// of cells.
					network.Grow(100, 20, 200)
				}
			}
		})

	}

	err = network.SaveToFile("network.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Failed saving network")
}

type processResult struct {
	succeeded   bool
	threadIndex int
}

/*
processLine fires this entire line in the neural network at once, hoping to get the desired output.

It will not add any synapses.
*/
func processLine(thread int, line string, network *potential.Network, vocab charrnn.Vocab, ch chan processResult) {
	lineChars := strings.Split(line, "")

	result := processResult{threadIndex: thread}
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
		for i := 0; i < 10; i++ {
			doneChan := make(chan bool)
			go func() {
				time.Sleep(potential.RefractoryPeriodMillis * 2 * time.Millisecond)
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
	fmt.Println(wasSuccessful, succeeded, "/", len(lineChars), "\n  ", line)
	result.succeeded = wasSuccessful
	ch <- result
}
