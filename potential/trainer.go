package potential

import (
	"fmt"
	"sync"
	"time"
)

const defaultWorkerThreads = 2
const initialNetworkNeurons = 200
const defaultNeuronSynapses = 5
const pretrainNeuronsToGrow = 20
const pretrainSynapsesToGrow = 50
const samplesBetweenPruningSessions = 20
const sleepBetweenInputTriggers = RefractoryPeriodMillis * time.Millisecond
const networkDisabledFizzleOutPeriod = 100 * time.Millisecond
const maxAllowedTimeForInputTriggeringOutput = synapseDelay * GrowPathExpectedMinimumSynapses * time.Millisecond

/*
TrainingSettings are
*/
type TrainingSettings struct {
	Threads uint
	// Data is the set of data structures that map the lowest units of the network onto input
	// and output cells.
	Data *Dataset
	// TrainingSamples is a list cell parings to fire for training. The cells must be
	// immortal and input or output cells.
	// Example: an array of lines, where each training sample is a letter and the one
	// that comes after it.
	TrainingSamples [][]*TrainingSample
}

/*
NewTrainingSettings sets up and returns a new TrainingSettings instance.
*/
func NewTrainingSettings() *TrainingSettings {
	d := Dataset{}
	dataset := &d
	settings := TrainingSettings{
		Threads:         defaultWorkerThreads,
		Data:            dataset,
		TrainingSamples: make([][]*TrainingSample, 0),
	}
	return &settings
}

/*
TrainingSample is a pairing of cells where the input should fire the output. It is just
for training, so theoretically one cell might fire another cell.
*/
type TrainingSample struct {
	InputCell  CellID
	OutputCell CellID
}

/*
PerceptionUnit is the smallest and core unit of a Dataset.
*/
type PerceptionUnit struct {
	Value      interface{}
	InputCell  CellID
	OutputCell CellID
}

/*
Dataset helps represent the smallest units of a trainable set of data, and maps each
unit so it can be used when training and sampling.

An example of data in a dataset would be the collection of letters from a character
neural network. Each letter would be mapped to a single input and output cell,
because the network uses groups of letters to predict groups of letters.

Another example might be inputs that are pixels coded by location and color. The outputs
could be categories of things that are in the photos, or that the photos represent.
*/
type Dataset struct {
	KeyToItem map[interface{}]PerceptionUnit
	CellToKey map[CellID]interface{}
}

/*
Trainer provides a way to run simulations on a neural network, then capture the results
and keep them, or re-strengthen the network.

The result of training is a new network which can be diffed onto an original network.
*/
type Trainer interface {
	PrepareData(*Network)
}

/*
Train executes the trainer's OnTrained method once complete.
*/
func Train(t Trainer, settings *TrainingSettings, network *Network) {
	t.PrepareData(network)

	// TODO: thread pool

	// p := pool.NewLimited(settings.Threads)
	// defer p.Close()

	var mux sync.Mutex
	totalTrainingPairs := len(settings.TrainingSamples)

	for i, batch := range settings.TrainingSamples {
		// train in parallel over this number of threads
		ch := make(chan Diff)

		net := CloneNetwork(network)
		net.GrowRandomNeurons(pretrainNeuronsToGrow, defaultNeuronSynapses)
		net.GrowRandomSynapses(pretrainSynapsesToGrow)
		go processBatch(batch, net, network, settings.Data, ch)

		diff := <-ch
		ApplyDiff(diff, network)

		fmt.Println("Round of lines done, line=", i, "/", totalTrainingPairs)
		if i%samplesBetweenPruningSessions == 0 {
			fmt.Println("Pruning...")
			fmt.Println("  before:", len(network.Cells), "cells,", len(network.Synapses), "synapses")
			mux.Lock()
			network.Prune()
			mux.Unlock()
			fmt.Println("  after:", len(network.Cells), "cells,", len(network.Synapses), "synapses")
		}

	}

}

/*
processBatch fires this entire line in the neural network at once, hoping to get the desired output.

It will not add any synapses.
*/
func processBatch(batch []*TrainingSample, network *Network, originalNetwork *Network, vocab *Dataset, ch chan Diff) {
	network.ResetForTraining()

	done := make(chan bool)

	successes := 0

	for _, ts := range batch {
		network.Cells[ts.InputCell].FireActionPotential()
		go time.AfterFunc(sleepBetweenInputTriggers, func() {
			done <- true
		})
		<-done
	}

	// wait for firings to go through the network
	go time.AfterFunc(maxAllowedTimeForInputTriggeringOutput, func() {
		for _, ts := range batch {
			if network.Cells[ts.OutputCell].WasFired {
				successes++
			}
		}
		done <- true
	})
	<-done

	wasSuccessful := successes == len(batch)

	network.Disabled = true

	var diff Diff

	// give the network some time to wind down
	time.AfterFunc(GrowPathExpectedMinimumSynapses*RefractoryPeriodMillis, func() {
		if wasSuccessful { // keep the training
			fmt.Println("  net fired all expected cells")
			diff = DiffNetworks(originalNetwork, network)
		} else {
			// We failed to generate the desired effect, so do a significant growth
			// of cells.
			fmt.Println("  net did not fire all cells, regrowing")
			for _, ts := range batch {
				network.GrowPathBetween(ts.InputCell, ts.OutputCell, GrowPathExpectedMinimumSynapses)
			}
			diff = DiffNetworks(originalNetwork, network)
		}
		done <- true
	})
	<-done
	ch <- diff
}
