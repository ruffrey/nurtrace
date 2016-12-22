package potential

import (
	"fmt"
	"math"
	"sync"
)

const defaultWorkerThreads = 2

/*
TrainingSettings are
*/
type TrainingSettings struct {
	Threads int
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

ONLY prune on the original network:

It is important to ONLY PRUNE on the original network. Diffing does not capture
what was removed - that is a surprising rabbit hole. So if you diff and prune on an
off-network, you can end up getting orphaned synapses - and bad integrity.

To effectively train in a distributed way, across machines, it may be necessary
to implement better diffing that does account for removed synapses and cells. But
for now this is more than adequate.
*/
func Train(t Trainer, settings *TrainingSettings, originalNetwork *Network) {
	// The next two are used to block until all threads are done and the function may return.
	var wg sync.WaitGroup
	done := make(chan bool)

	// The next three are used to synchronize 1) merging of thread networks onto original,
	// and 2) cloning of original network. Otherwise things get quickly out of sync, and
	// race conditions cause nil pointer reference issues when network is overwritten with
	// the latest originalNetwork.
	var origNetCloneMux sync.Mutex              // only clone networks inside this mutex
	chNetworkSync := make(chan *Network, 1)     // blocking channel for sending net to be merged
	chNetworkSyncCallback := make(chan bool, 1) // blocking channel waiting for response of net merge

	// precheck
	ok, report := CheckIntegrity(originalNetwork)
	if !ok {
		fmt.Println(report)
		panic("integrity failed before training")
	}
	t.PrepareData(originalNetwork)

	// DO NOT PRUNE on the cloned networks! See not in method comment above.
	lenAllSamples := len(settings.TrainingSamples)
	partSize := math.Ceil(float64(lenAllSamples) / float64(settings.Threads))
	fmt.Println(partSize, "samples per thread")
	for thread := 0; thread < settings.Threads; thread++ {
		wg.Add(1)
		go func(thread int) {
			network := CloneNetwork(originalNetwork)
			from := int(float64(thread) * partSize)
			to := int(float64(thread+1) * partSize)
			fmt.Println("thread", thread, "from", from, "to", to)
			samples := settings.TrainingSamples[from:to]
			for i, batch := range samples {
				createdEffect, diff := processBatch(batch, network, settings.Data)

				if !createdEffect {
					// if ok, report := CheckIntegrity(originalNetwork); !ok {
					// 	fmt.Println("ApplyDiff: network has no integrity BEFORE")
					// 	report.Print()
					// 	panic("no integrity")
					// }
					ApplyDiff(diff, network)
					// if ok, report := CheckIntegrity(originalNetwork); !ok {
					// 	fmt.Println("ApplyDiff: network has no integrity AFTER")
					// 	diff.Print()
					// 	report.Print()
					// 	panic("no integrity")
					// }
				}

				if i%samplesBetweenMergingSessions == 0 {
					if i == 0 { // do not allow diffing/pruning before getting started!
						continue
					}
					origNetCloneMux.Lock()
					chNetworkSync <- network
					<-chNetworkSyncCallback
					// after merging back changes, get up to date with the latest network
					network = CloneNetwork(originalNetwork)
					origNetCloneMux.Unlock()
				}

			}

			fmt.Println("Network on thread", thread, "done")
			chNetworkSync <- network
			<-chNetworkSyncCallback
			fmt.Println("Applied diff on thread", thread)
			wg.Done()
		}(thread)
	}

	go func() {
		wg.Wait()
		done <- true
	}()

	var mergeNum int
	for {
		select {
		case network := <-chNetworkSync:
			oDiff := DiffNetworks(originalNetwork, network)
			// if ok, report := CheckIntegrity(originalNetwork); !ok {
			// 	fmt.Println("ApplyDiff: originalNetwork has no integrity BEFORE")
			// 	report.Print()
			// 	panic("no integrity")
			// }
			ApplyDiff(oDiff, originalNetwork)
			// if ok, report := CheckIntegrity(originalNetwork); !ok {
			// 	fmt.Println("ApplyDiff: originalNetwork has no integrity AFTER")
			// 	diff.Print()
			// 	report.Print()
			// 	panic("no integrity")
			// }

			mergeNum++
			if mergeNum%settings.Threads == 0 {
				// DO NOT prune on the one-off network that has not been merged back to main.
				// if ok, report := CheckIntegrity(network); !ok {
				// 	fmt.Println("Prune: network has no integrity BEFORE pruning")
				// 	report.Print()
				// 	panic("no integrity")
				// }
				originalNetwork.Prune()
				// if ok, report := CheckIntegrity(network); !ok {
				// 	fmt.Println("Prune: network has no integrity AFTER pruning")
				// 	report.Print()
				// 	fmt.Println("intended to remove synapses:", synapsesToRemove)
				// 	for cellID, synapseID := range report.cellHasMissingAxonSynapse {
				// 		fmt.Println("  cellHasMissingAxonSynapse", cellID, network.Cells[cellID])
				// 		fmt.Println("  cellHasMissingAxonSynapse", synapseID, network.Synapses[synapseID])
				// 	}
				// 	for cellID, synapseID := range report.cellHasMissingDendriteSynapse {
				// 		fmt.Println("  cellHasMissingDendriteSynapse", cellID, network.Cells[cellID])
				// 		fmt.Println("  cellHasMissingDendriteSynapse", synapseID, network.Synapses[synapseID])
				// 	}
				// 	panic("no integrity")
				// }
				fmt.Println("Progress:",
					math.Floor(((float64(mergeNum)*float64(samplesBetweenMergingSessions))/float64(lenAllSamples))*100), "%")
			}
			chNetworkSyncCallback <- true
		case <-done:
			return
		}
	}

	// TODO: final prune?
}

/*
processBatch fires this entire line in the neural network at once, hoping to get the desired output.

It will not add any synapses.
*/
func processBatch(batch []*TrainingSample, originalNetwork *Network, vocab *Dataset) (wasSuccessful bool, diff Diff) {
	network := CloneNetwork(originalNetwork)
	network.ResetForTraining()

	successes := 0

	for _, ts := range batch {
		// cell, ok := network.Cells[ts.InputCell]
		// if !ok {
		// 	fmt.Println("error: input cell missing from network", ts)
		// }
		// cell.FireActionPotential()
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
	// see how many of the cells we wanted actually fired
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
		// of cells. Grow some random stuff to introduce a little noise and new
		// things to grab onto.
		network.GrowRandomNeurons(retrainNeuronsToGrow, defaultNeuronSynapses)
		network.GrowRandomSynapses(retrainRandomSynapsesToGrow)

		// (re)grow paths between each expected input and output.
		for _, ts := range batch {
			network.GrowPathBetween(ts.InputCell, ts.OutputCell, GrowPathExpectedMinimumSynapses)
		}
		diff = DiffNetworks(originalNetwork, network)
	}

	return wasSuccessful, diff
}
