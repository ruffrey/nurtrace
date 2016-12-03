package main

import (
	"bleh/charrnn"
	"bleh/potential"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

const fireCharacterIterations = 4
const initialNetworkNeurons = 2000
const defaultNeuronSynapses = 10
const pretrainNeuronsToGrow = 20
const pretrainSynapsesToGrow = 100
const growPathExpectedSynapses = 20
const sleepBetweenInputTriggers = potential.RefractoryPeriodMillis * time.Millisecond
const threads = 1
const networkDisabledFizzleOutPeriod = 100 * time.Millisecond

var samples = flag.Int("sample", 0, "Pass this flag with seed to do a sample instead of training")
var seed = flag.String("seed", "", "Seeds the neural network when sampling")

func sample(vocab charrnn.Vocab, network *potential.Network) {
	network.Disabled = false
	network.ResetForTraining()
	fmt.Print("\n")
	seedChars := strings.Split(*seed, "")

	go func() {
		for _, v := range vocab.CharToItem {
			network.Cells[v.OutputCell].OnFired = append(
				network.Cells[v.OutputCell].OnFired,
				func(cell potential.CellID) {
					s := vocab.CellToChar[cell]
					if s == "END" {
						fmt.Print("\n")
						os.Exit(0)
					}
					if s == "START" {
						return
					}
					fmt.Print(s)
				},
			)
		}
	}()

	for {
		for _, char := range seedChars {
			for i := 0; i < 4; i++ {
				ch := make(chan bool)
				time.AfterFunc(sleepBetweenInputTriggers, func() {
					network.Cells[vocab.CharToItem[char].InputCell].FireActionPotential()
					ch <- true
				})
				<-ch
			}
		}
	}
}

func main() {
	// defer profile.Start(profile.MemProfile).Stop()
	// defer profile.Start(profile.CPUProfile).Stop()
	// doTrace()

	// start by initializing the network from disk or whatever
	var network *potential.Network
	var err error
	network, err = potential.LoadNetworkFromFile("network.json")
	if err != nil {
		fmt.Println("Unable to load network from file; creating new one.", err)
		newN := potential.NewNetwork()
		network = &newN
		neuronsToAdd := initialNetworkNeurons
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
		fmt.Println("Loaded network from disk,", len(network.Cells), "cells",
			len(network.Synapses), "synapses")
	}

	bytes, err := ioutil.ReadFile("short.txt")
	if err != nil {
		panic(err)
	}

	text := string(bytes)
	lines := strings.Split(text, "\n")

	vocab := charrnn.NewVocab(text, network)

	fmt.Println("Loaded vocab, length=", len(vocab.CharToItem))

	flag.Parse()

	if *samples > 0 && len(*seed) > 0 {
		fmt.Println("Sampling", *samples, "characters with text: ", *seed)
		sample(vocab, network)
		return
	}

	network.Disabled = true // we just will never need it to fire

	fmt.Println("Beginning training", threads, "simultaneous sessions")

	totalLines := len(lines)
	for i := 0; i < totalLines; {
		var wg sync.WaitGroup

		// train in parallel over this number of threads
		ch := make(chan potential.Diff)

		networkCopies := make(map[int]*potential.Network)
		results := make(map[int]potential.Diff)
		for thread := 0; thread < threads; thread++ {
			if i >= totalLines { // ran out of lines
				break
			}

			line := lines[i]
			net := potential.CloneNetwork(network)
			net.GrowRandomNeurons(pretrainNeuronsToGrow, defaultNeuronSynapses)
			net.GrowRandomSynapses(pretrainSynapsesToGrow)
			// fmt.Println("starting thread", thread)
			wg.Add(1)
			go func(net *potential.Network, thread int) {
				processLine(thread, line, net, network, vocab, ch)
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
		for diff := range ch {
			results[diff.Worker] = diff
		}

		// now that all threads are finished, read their results and modify the network in
		// series

		// End all network firings, let them finish, then do diffing or growing.

		done := make(chan bool)
		time.AfterFunc(networkDisabledFizzleOutPeriod, func() {
			// TODO: move this back down into processLine and have it return a diff instead
			// of results
			for thread := 0; thread < threads; thread++ {
				diff := results[thread]
				potential.ApplyDiff(diff, network)
			}
			done <- true
		})
		<-done
		fmt.Println("Round of lines done, line=", i, "/", totalLines)
	}

	fmt.Println(len(network.Cells), "cells", len(network.Synapses), "synapses")
	fmt.Println("Pruning...")
	network.Prune()
	fmt.Println(len(network.Cells), "cells", len(network.Synapses), "synapses")

	err = network.SaveToFile("network.json")
	if err != nil {
		fmt.Println("Failed saving network")
		fmt.Println(err)
		return
	}
	fmt.Println("Done.")
}

/*
processLine fires this entire line in the neural network at once, hoping to get the desired output.

It will not add any synapses.
*/
func processLine(thread int, line string, network *potential.Network, originalNetwork *potential.Network, vocab charrnn.Vocab, ch chan potential.Diff) {
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
			inChar = vocab.CharToItem["START"]
			outChar = vocab.CharToItem[char]
		} else if isLast {
			inChar = vocab.CharToItem[char]
			outChar = vocab.CharToItem["END"]
		} else {
			inChar = vocab.CharToItem[lineChars[ix-1]]
			outChar = vocab.CharToItem[char]
		}

		network.ResetForTraining()

		for i := 0; i < fireCharacterIterations; i++ {
			doneChan := make(chan bool)
			go func(i int) {
				time.AfterFunc(sleepBetweenInputTriggers, func() {
					network.Cells[inChar.OutputCell].FireActionPotential()
					doneChan <- true
				})
			}(i)
			<-doneChan
		}

		if network.Cells[outChar.InputCell].WasFired {
			succeeded++
		}
	}
	wasSuccessful := succeeded == len(lineChars)
	// fmt.Println(wasSuccessful, succeeded, "/", len(lineChars), "\n  ", line)
	fmt.Println("disabling network=", thread)
	network.Disabled = true

	var diff potential.Diff

	done := make(chan bool)
	time.AfterFunc(10*potential.RefractoryPeriodMillis, func() {

		if wasSuccessful { // keep the training
			fmt.Println("Keep diff for thread=", thread)
			diff = potential.DiffNetworks(originalNetwork, network)
		} else {
			// We failed to generate the desired effect, so do a significant growth
			// of cells.
			fmt.Println("Discard diff for thread=", thread, "and regrow")
			for _, vocabItem := range vocab.CharToItem {
				network.GrowPathBetween(vocabItem.InputCell, vocabItem.OutputCell, growPathExpectedSynapses)
			}
			diff = potential.DiffNetworks(originalNetwork, network)

		}
		diff.Worker = thread

		done <- true
	})
	<-done
	ch <- diff
}
