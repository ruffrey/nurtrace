package potential

import (
	"bleh/laws"
	"encoding/gob"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

/*
TrainingSettings are
*/
type TrainingSettings struct {
	// Data is the set of data structures that map the lowest units of the network onto input
	// and output cells.
	Data *Dataset
	// TrainingSamples is a list cell parings to fire for training. The cells must be
	// immortal and input or output cells.
	// Example: an array of lines, where each training sample is a letter and the one
	// that comes after it.
	TrainingSamples [][]*TrainingSample

	Threads    int
	Workerfile string
}

/*
NewTrainingSettings sets up and returns a new TrainingSettings instance.
*/
func NewTrainingSettings() *TrainingSettings {
	d := Dataset{}
	dataset := &d
	settings := TrainingSettings{
		Threads:         runtime.NumCPU(),
		Data:            dataset,
		TrainingSamples: make([][]*TrainingSample, 0),
	}
	return &settings
}

func copySettingsWithNewSamples(originalSettings *TrainingSettings, samples [][]*TrainingSample) *TrainingSettings {
	settings := NewTrainingSettings()
	settings.Data = originalSettings.Data
	settings.TrainingSamples = samples
	return settings
}

/*
LoadTrainingSettingsFromFile loads the training settings from a file on disk.
*/
func LoadTrainingSettingsFromFile(localFilepath string) (*TrainingSettings, error) {
	settings := NewTrainingSettings()

	file, err := os.Open(localFilepath)
	if err != nil {
		return settings, err
	}
	// Create a decoder and receive a value.
	dec := gob.NewDecoder(file)
	err = dec.Decode(&settings)
	if err != nil {
		log.Fatal("decode:", err)
	}
	settings.Threads = runtime.NumCPU() // use number of threads on the remote machine
	return settings, err
}

/*
SaveTrainingSettingsToFile dumps the training settings as json to disk.
*/
func SaveTrainingSettingsToFile(settings *TrainingSettings, localFilepath string) error {
	file, err := os.Create(localFilepath)
	if err != nil {
		return err
	}
	// Create an encoder and send a value.
	enc := gob.NewEncoder(file)
	err = enc.Encode(settings)
	if err != nil {
		log.Fatal("encode:", err)
	}

	if err != nil {
		return err
	}
	return nil
}

// use the current time for easy reading, but also generate a random token
func randFilename(prefix string, ext string) string {
	now := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	return fmt.Sprintf("%s_%s_%d.%s", prefix, now, rand.Uint32(), ext)
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
Train runs the training samples on local and remote threads, and applies them to
the originalNetwork.

ONLY prune on the original network:

It is important to ONLY PRUNE on the original network. Diffing does not capture
what was removed - that is a surprising rabbit hole. So if you diff and prune on an
off-network, you can end up getting orphaned synapses - and bad integrity.

To effectively train in a distributed way, across machines, it may be necessary
to implement better diffing that does account for removed synapses and cells. But
for now this is more than adequate.
*/
func Train(settings *TrainingSettings, originalNetwork *Network, isRemoteWorkerWithTag string) {
	shouldPrune := isRemoteWorkerWithTag == ""
	if shouldPrune {
		isRemoteWorkerWithTag = "<local>"
	}
	// The next two are used to block until all threads are done and the function may return.
	var wg sync.WaitGroup
	done := make(chan bool)
	// list of remote worker server locations that we will ssh into
	var remoteWorkers []string
	var remoteWorkerWeights []int
	var remoteWorkerTotalWeights int
	var err error
	// The next three are used to synchronize 1) merging of thread networks onto original,
	// and 2) cloning of original network. Otherwise things get quickly out of sync, and
	// race conditions cause nil pointer reference issues when network is overwritten with
	// the latest originalNetwork.
	var origNetCloneMux sync.Mutex              // only clone networks inside this mutex
	chNetworkSync := make(chan *Network, 1)     // blocking channel for sending net to be merged
	chNetworkSyncCallback := make(chan bool, 1) // blocking channel waiting for response of net merge

	if settings.Workerfile != "" {
		remoteWorkers, remoteWorkerWeights, remoteWorkerTotalWeights, err = readWorkerfile(settings.Workerfile)
		if err != nil {
			panic(err)
		}
	}

	// precheck
	ok, report := CheckIntegrity(originalNetwork)
	if !ok {
		fmt.Println(isRemoteWorkerWithTag, report)
		panic("integrity failed before training")
	}

	// DO NOT PRUNE on the cloned networks! See note in method comment above.
	lenAllSamples := len(settings.TrainingSamples)
	jobChunks := settings.Threads + remoteWorkerTotalWeights
	partSize := math.Ceil(float64(lenAllSamples) / float64(jobChunks))
	maxSampleIndex := len(settings.TrainingSamples) - 1
	fmt.Println(isRemoteWorkerWithTag, partSize,
		"samples per chunk,", settings.Threads, "local threads,",
		remoteWorkerTotalWeights, "remote weights (", jobChunks, "chunks )")
	sampleCursor := 0
	threadIteration := 0
	for thread := 0; threadIteration < jobChunks; thread++ {
		wg.Add(1)
		isRemote := thread < len(remoteWorkers)

		if isRemote {
			threadIteration += remoteWorkerWeights[thread]
		} else {
			threadIteration++
		}
		to := threadIteration * int(partSize)
		// protect likely array out of bounds on last thread
		if to > maxSampleIndex {
			to = maxSampleIndex
		}
		fmt.Println(sampleCursor, to)
		samples := settings.TrainingSamples[sampleCursor:to]

		if isRemote {
			fmt.Println(isRemoteWorkerWithTag, "thread", thread, "from",
				sampleCursor, "to", to, "on", remoteWorkers[thread])
		} else {
			fmt.Println(isRemoteWorkerWithTag, "thread", thread, "from",
				sampleCursor, "to", to)
		}

		sampleCursor = to

		origNetCloneMux.Lock()
		network := CloneNetwork(originalNetwork)
		origNetCloneMux.Unlock()

		go func(thread int) {

			// first we start the remote workers
			if isRemote {
				w, err := NewWorker(remoteWorkers[thread])
				if err != nil {
					fmt.Println("error making new worker", remoteWorkers[thread])
					panic(err)
				}
				defer w.conn.Close()
				w.TranserExecutable()
				copiedSettings := copySettingsWithNewSamples(settings, samples)
				tempSettingsFile := randFilename("settings", "gob")
				tempNetworkFile := randFilename("network", "json")
				err = SaveTrainingSettingsToFile(copiedSettings, tempSettingsFile)
				if err != nil {
					panic(err)
				}
				err = network.SaveToFile(tempNetworkFile)
				if err != nil {
					panic(err)
				}
				network, err = w.Train(tempSettingsFile, tempNetworkFile)
				if err != nil {
					fmt.Println("Error from remote worker", w.host)
					panic(err)
				}
				fmt.Println("Remote thread", thread, w.host, "done")
				chNetworkSync <- network
				<-chNetworkSyncCallback
				fmt.Println("Applied final diff on remote thread", thread)
				wg.Done()
				return
			}

			// normal local worker
			for i, batch := range samples {
				createdEffect, diff := processBatch(batch, network, settings.Data)

				if !createdEffect {
					ApplyDiff(diff, network)
				}

				if i%laws.SamplesBetweenMergingSessions == 0 {
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

			fmt.Println(isRemoteWorkerWithTag, "local thread", thread, "done")
			chNetworkSync <- network
			<-chNetworkSyncCallback
			fmt.Println(isRemoteWorkerWithTag, "applied final diff on local thread", thread)
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
			ApplyDiff(oDiff, originalNetwork)

			mergeNum++
			if mergeNum%settings.Threads == 0 {
				// DO NOT prune on the one-off network that has not been merged back to main.
				if shouldPrune {
					// originalNetwork.Prune()
					dupes := findDupeSynapses(originalNetwork)
					for _, dupeGroup := range dupes {
						dedupeSynapses(dupeGroup, originalNetwork)
					}
				}
				fmt.Println(isRemoteWorkerWithTag, "Progress:",
					math.Floor(((float64(mergeNum)*float64(laws.SamplesBetweenMergingSessions))/float64(lenAllSamples))*100), "%,")
				network.PrintTotals()
			}
			chNetworkSyncCallback <- true
		case <-done:
			return
		}
	}

}

/*
processBatch fires this entire line in the neural network at once, hoping to get the desired output.

It will not add any synapses.
*/
func processBatch(batch []*TrainingSample, originalNetwork *Network, data *Dataset) (wasSuccessful bool, diff Diff) {
	firedOutputBatches := make([]*TrainingSample, 0)
	unfiredOutputBatches := make([]*TrainingSample, 0)
	expectedOutputCells := make(map[CellID]bool)
	allInputCells := make(map[CellID]bool)
	noisyOutputCells := make(map[CellID]bool)
	network := CloneNetwork(originalNetwork)

	network.ResetForTraining()

	for _, ts := range batch {
		expectedOutputCells[ts.OutputCell] = true // need these later to know which cells should not have fired
		allInputCells[ts.InputCell] = true

		network.Cells[ts.InputCell].FireActionPotential()
		// TODO: should we step here? or not?
		network.Step()
	}

	// give firings time to travel through the network
	for i := 0; i < laws.GrowPathExpectedMinimumSynapses; i++ {
		hasMore := network.Step()
		if !hasMore {
			break
		}
	}
	// see how many of the cells we wanted actually fired
	for _, ts := range batch {
		if network.Cells[ts.OutputCell].WasFired {
			firedOutputBatches = append(firedOutputBatches, ts)
		} else {
			unfiredOutputBatches = append(unfiredOutputBatches, ts)
		}
	}
	// which cells should not have fired
	for _, unit := range data.KeyToItem {
		if _, wasExpected := expectedOutputCells[unit.OutputCell]; !wasExpected && network.Cells[unit.OutputCell].WasFired {
			noisyOutputCells[unit.OutputCell] = true
		}
	}

	// wind down the network
	network.Disabled = true
	hasMore := network.Step()

	if hasMore {
		fmt.Println("warn: more cell firings existed after disabling network and stepping")
	}

	firedAllExpected := len(unfiredOutputBatches) == 0
	firedNoNoise := len(noisyOutputCells) == 0
	wasSuccessful = firedAllExpected && firedNoNoise

	if !wasSuccessful {
		// most critical part of training, ensure we backprop and whatnot

		// merge all the good synapses into a single map
		goodSynapses := make(map[SynapseID]bool)
		for _, ts := range firedOutputBatches {
			gs := backwardTraceFirings(network, ts.OutputCell, ts.InputCell)
			for synapseID := range gs {
				goodSynapses[synapseID] = true
			}
		}
		if !firedNoNoise {
			backwardTraceNoiseAndInhibit(network, allInputCells, noisyOutputCells, goodSynapses)
		}

		// reinforce all good synapses. These are synapses that fire when one of the
		// input cells fires.
		for synapseID := range goodSynapses {
			synapse := network.getSyn(synapseID)
			synapse.reinforce()
		}

		// We failed to generate the desired effect, so do a growth of cells.
		// Grow some random stuff to introduce a little noise and new
		// things to grab onto. Not sure if it is a good idea or not.
		// Maybe instead, GrowPathBetween should add cells.
		network.GrowRandomNeurons(laws.RetrainNeuronsToGrow, laws.DefaultNeuronSynapses)

		// (re)grow paths between each expected input and output,
		// but only when the inputs did not fire all outputs.
		if !firedAllExpected {
			for _, ts := range unfiredOutputBatches {
				network.GrowPathBetween(ts.InputCell, ts.OutputCell, laws.GrowPathExpectedMinimumSynapses)
			}
		}

		diff = DiffNetworks(originalNetwork, network)
	}

	return wasSuccessful, diff
}
