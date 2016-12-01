package main

import (
	"bleh/charrnn"
	"bleh/potential"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"
)

func main() {
	threads := 1

	// start by initializing the network from disk or whatever
	var network *potential.Network
	var err error
	n, err := potential.LoadNetworkFromFile("network.json")
	if err != nil {
		fmt.Println("No existing network in file; creating new one.", err)
		newN := potential.NewNetwork()
		network = &newN
		neuronsToAdd := 2000
		defaultNeuronSynapses := 10
		synapsesToAdd := 0
		network.Grow(neuronsToAdd, defaultNeuronSynapses, synapsesToAdd)
		fmt.Println("Created network")
		fmt.Println("Saving to disk")
		// err = network.SaveToFile("network.json")
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }
	} else {
		network = &n
		fmt.Println("Loaded network from disk,", len(network.Cells), "cells")
	}

	network.Disabled = true // we just will never need it to fire

	bytes, err := ioutil.ReadFile("short.txt")
	if err != nil {
		panic(err)
	}

	text := string(bytes)
	lines := strings.Split(text, "\n")

	vocab := charrnn.NewVocab(text, network)

	fmt.Println("Loaded vocab, length=", len(vocab))

	fmt.Println("Beginning training", threads, "simultaneous sessions")

	totalLines := len(lines)
	for i := 0; i < totalLines; {
		var wg sync.WaitGroup

		// train in parallel over this number of threads
		ch := make(chan processResult)

		networkCopies := make(map[int]*potential.Network)
		results := make(map[int]processResult)
		for thread := 0; thread < threads; thread++ {
			if i >= totalLines { // ran out of lines
				break
			}

			line := lines[i]
			net := potential.CloneNetwork(network)
			net.GrowRandomNeurons(20, 10)
			net.GrowRandomSynapses(100)
			// fmt.Println("starting thread", thread)
			wg.Add(1)
			go func(net *potential.Network, thread int) {
				processLine(thread, line, net, vocab, ch)
				wg.Done()
			}(net, thread)

			networkCopies[thread] = net

			i++

		}

		go func() {
			wg.Wait()
			close(ch)
		}()

		// read results from the threads as they come in
		for result := range ch {
			results[result.threadIndex] = result
		}

		// now that all threads are finished, read their results and modify the network in
		// series

		// End all network firings, let them finish, then do diffing or growing.
		for ix, net := range networkCopies {
			fmt.Println("disabling network=", ix)
			net.Disabled = true
		}

		done := make(chan bool)
		time.AfterFunc(100*time.Millisecond, func() {
			for thread := 0; thread < threads; thread++ {
				r := results[thread]
				net := networkCopies[r.threadIndex]
				if r.succeeded { // keep the training
					diff := potential.DiffNetworks(network, net)
					fmt.Println("applying natural diff for thread=", thread)
					potential.ApplyDiff(diff, network)
				} else {
					// We failed to generate the desired effect, so do a significant growth
					// of cells.
					fmt.Println("disgarding growth for thread=", thread, "and adding more connections")
					for _, vocabItem := range vocab {
						// finish := make(chan bool)
						// go func() {
						net.GrowPathBetween(vocabItem.InputCell, vocabItem.OutputCell, 20)
						// finish <- true
						// }()
						// <-finish
					}
					diff := potential.DiffNetworks(network, net)
					fmt.Println("applying growth diff for thread=", thread)
					potential.ApplyDiff(diff, network)
				}
			}
			done <- true
		})
		<-done
		fmt.Println("Round of lines done, line=", i, "/", totalLines)
	}

	err = network.SaveToFile("network.json")
	if err != nil {
		fmt.Println("Failed saving network")
		fmt.Println(err)
		return
	}
	fmt.Println("Reached end")
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
		sleepTime := (potential.RefractoryPeriodMillis + 1) * time.Millisecond

		for i := 0; i < 10; i++ {
			doneChan := make(chan bool)
			go func(i int) {
				time.Sleep(sleepTime)
				network.Cells[inChar.OutputCell].FireActionPotential()
				doneChan <- true
			}(i)
			<-doneChan
		}

		if network.Cells[outChar.InputCell].WasFired {
			succeeded++
		}
	}
	wasSuccessful := succeeded == len(lineChars)
	// fmt.Println(wasSuccessful, succeeded, "/", len(lineChars), "\n  ", line)
	result.succeeded = wasSuccessful
	ch <- result
}
