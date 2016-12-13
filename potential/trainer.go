package potential

import (
	"fmt"
	"sync"
	"time"
)

const defaultWorkerThreads = 2
const initialNetworkNeurons = 200
const defaultNeuronSynapses = 5
const pretrainNeuronsToGrow = 10
const pretrainSynapsesToGrow = 20
const samplesBetweenPruningSessions = 16
const defaultSynapseMinFireThreshold = 8
const networkDisabledFizzleOutPeriod = 100 * time.Millisecond

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
func Train(t Trainer, settings *TrainingSettings, originalNetwork *Network) {
	var mux sync.Mutex
	var wg sync.WaitGroup

	t.PrepareData(originalNetwork)

	for thread := uint(0); thread < settings.Threads; thread++ {
		wg.Add(1)
		go func(thread uint) {
			network := CloneNetwork(originalNetwork)

			// totalTrainingPairs := len(settings.TrainingSamples)

			for i, batch := range settings.TrainingSamples {
				net := CloneNetwork(network)
				createdEffect, diff := processBatch(batch, net, network, settings.Data)

				if !createdEffect {
					ApplyDiff(diff, network)
				}

				if i%samplesBetweenPruningSessions == 0 {
					if i == 0 { // do not prune before getting started!
						continue
					}
					// beforeCells := len(network.Cells)
					// beforeSynapses := len(network.Synapses)
					network.Prune()
					// afterCells := len(network.Cells)
					// afterSynapses := len(network.Synapses)
					// fmt.Println("Prune", thread, i, "/", totalTrainingPairs, " cells=", afterCells, "synapses=", afterSynapses, "removed cells=", beforeCells-afterCells, "removed synapses=", beforeSynapses-afterSynapses)
				}

			}

			fmt.Println("Network on thread", thread, "done")
			mux.Lock()
			finalDiff := DiffNetworks(originalNetwork, network)
			ApplyDiff(finalDiff, originalNetwork)
			mux.Unlock()
			fmt.Println("Applied diff on thread", thread)
			wg.Done()
		}(thread)
	}

	wg.Wait()
}

/*
processBatch fires this entire line in the neural network at once, hoping to get the desired output.

It will not add any synapses.
*/
func processBatch(batch []*TrainingSample, network *Network, originalNetwork *Network, vocab *Dataset) (wasSuccessful bool, diff Diff) {
	network.ResetForTraining()

	successes := 0

	for _, ts := range batch {
		network.Cells[ts.InputCell].FireActionPotential()
		// TODO: should we step here? or not?
		network.Step()
	}

	// give for firings time to go through the network
	for i := 0; i < GrowPathExpectedMinimumSynapses; i++ {
		hasMore := network.Step()
		if !hasMore {
			break
		}
	}
	for _, ts := range batch {
		if network.Cells[ts.OutputCell].WasFired {
			successes++
		}
	}

	wasSuccessful = successes == len(batch)

	// wind down the network
	network.Disabled = true
	hasMore := network.Step()

	if hasMore {
		fmt.Println("warn: more cell firings existed after disabling network and stepping")
	}

	if !wasSuccessful {
		// We failed to generate the desired effect, so do a significant growth
		// of cells.
		// fmt.Println("  net did not fire all cells, regrowing")
		// grow some random stuff
		network.GrowRandomNeurons(pretrainNeuronsToGrow, defaultNeuronSynapses)
		network.GrowRandomSynapses(pretrainSynapsesToGrow)

		for _, ts := range batch {
			// grow paths between synapses, too
			// fmt.Println("  post-train diff adding synapses for", ts.InputCell, ts.OutputCell)
			network.GrowPathBetween(ts.InputCell, ts.OutputCell, GrowPathExpectedMinimumSynapses)
			// fmt.Println("    added", len(sEnd)+len(sAdded), "synapses")
		}
		diff = DiffNetworks(originalNetwork, network)
	} else {
		// fmt.Println("  net fired all expected cells, no changes")
	}

	return wasSuccessful, diff
}
