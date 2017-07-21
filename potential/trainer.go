package potential

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

func copyVocabWithNewSamples(original *Vocabulary, samples []sample) *Vocabulary {
	newVocab := NewVocabulary(CloneNetwork(original.Net))
	// copy Inputs and Outputs maps (maps pass by reference so new ones are needed)
	for k, v := range original.Inputs {
		newVocab.Inputs[k] = &VocabUnit{
			Value:      v.Value,
			InputCells: cloneFiringPattern(v.InputCells),
		}
	}
	for k, v := range original.Outputs {
		newVocab.Outputs[k] = &OutputCollection{
			Value:       v.Value,
			FirePattern: cloneFiringPattern(v.FirePattern),
		}
	}
	newVocab.Threads = original.Threads
	newVocab.Workerfile = original.Workerfile
	// this is the different one
	newVocab.Samples = samples
	return newVocab
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

TODO: Evaulate whether it still makes sense to have a separate InputCell
and OutputCell.
This might map onto the use case of generating text char->text char, but it
does not seem to fit the use case of generating other kinds of data, especially
when the input data is not the same kind of data as the output.
It may be that, due to refactoring over time, we don't even need separate InputCell
and OutputCell props for charrnn.
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
the original network.

ONLY dedupe on the original network on a single protected thread.

Diffing does not capture what was removed - that is a surprising rabbit hole.
So if you diff and dedupe on an off-network, you can end up getting orphaned
synapses - and bad integrity.

The Inputs should already be setup, before training. However the Outputs
will change, so they should be merged along with the network merge.
*/
func Train(masterVocab *Vocabulary, isRemoteWorkerWithTag string) {
	// TODO: deduping is turned off because of #40
	shouldDedupe := true
	// shouldDedupe := isRemoteWorkerWithTag == ""
	if shouldDedupe {
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
	chSynchVocab := make(chan *Vocabulary)
	chSendBackVocab := make(chan *Vocabulary)

	if masterVocab.Workerfile != "" {
		remoteWorkers, remoteWorkerWeights, remoteWorkerTotalWeights, err = readWorkerfile(masterVocab.Workerfile)
		if err != nil {
			panic(err)
		}
	}

	// precheck
	ok, report := CheckIntegrity(masterVocab.Net)
	if !ok {
		log.Println(isRemoteWorkerWithTag, report)
		panic("integrity failed before training")
	}

	// Preparing samles for each worker/local and each thread
	lenAllSamples := len(masterVocab.Samples)
	jobChunks := masterVocab.Threads + remoteWorkerTotalWeights
	partSize := math.Ceil(float64(lenAllSamples) / float64(jobChunks))
	log.Println(isRemoteWorkerWithTag, partSize,
		"samples per chunk,", masterVocab.Threads, "local threads,",
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
		to := threadIteration*int(partSize) + 1
		// protect likely array out of bounds on last thread
		if to > lenAllSamples {
			to = lenAllSamples
		}
		samples := masterVocab.Samples[sampleCursor:to]
		vocab := copyVocabWithNewSamples(masterVocab, samples)

		if isRemote {
			log.Println(isRemoteWorkerWithTag, "thread", thread, "from",
				sampleCursor, "to", to, "on", remoteWorkers[thread])
		} else {
			log.Println(isRemoteWorkerWithTag, "thread", thread, "from",
				sampleCursor, "to", to)
		}

		sampleCursor = to

		go func(thread int) {

			// first we start the remote workers
			if isRemote {
				w, err := NewWorker(remoteWorkers[thread])
				if err != nil {
					log.Println("error making new worker", remoteWorkers[thread])
					panic(err)
				}
				defer w.conn.Close()
				w.TranserExecutable()
				tempVocabFile := randFilename("vocab", "json")
				tempNetworkFile := randFilename("network", "nur")
				err = vocab.SaveToFile(tempVocabFile)
				if err != nil {
					panic(err)
				}
				err = vocab.Net.SaveToFile(tempNetworkFile)
				if err != nil {
					panic(err)
				}
				vocab, err = w.Train(tempVocabFile, tempNetworkFile)
				if err != nil {
					log.Println("Error from remote worker", w.host)
					panic(err)
				}
				log.Println("Remote thread", thread, w.host, "done")
				chSynchVocab <- vocab
				log.Println("Applied final diff on remote thread", thread)
				wg.Done()
				return
			}

			// normal local worker
			thisTag := isRemoteWorkerWithTag + "<" + strconv.Itoa(thread) + ">"
			RunFiringPatternTraining(vocab, chSynchVocab, chSendBackVocab, thisTag)

			log.Println(isRemoteWorkerWithTag,
				"local thread", thread, "done")
			log.Println(isRemoteWorkerWithTag,
				"applied final diff on local thread", thread)
			wg.Done()
		}(thread)
	}

	go func() {
		wg.Wait()
		done <- true
	}()

	merges := 0
	for {
		select {
		case vocab := <-chSynchVocab:
			oDiff := DiffNetworks(masterVocab.Net, vocab.Net)
			idChanges := ApplyDiff(oDiff, masterVocab.Net)

			// IDs that were reassigned need to be updated throughout
			// vocab also. Merging a diff on the network does not handle that.
			if len(idChanges) > 0 {
				rerouteChangedIOCellIDs(idChanges, vocab) // old vocab
			}

			mergeAllOutputs(masterVocab.Outputs, vocab.Outputs)
			// should not be adding inputs on any network except the main one.
			// it presents a complication merging the firing patterns
			// from the branched network when the firing patterns reference
			// cells that do not yet exist on the new network
			mergeAllInputs(masterVocab.Inputs, vocab.Inputs)

			merges++
			if merges%(masterVocab.Threads+1) == 0 {
				if shouldDedupe {
					dupes := findDupeSynapses(masterVocab.Net)
					for _, dupeGroup := range dupes {
						dedupeSynapses(dupeGroup, masterVocab.Net)
					}
				}
				masterVocab.Net.PrintTotals()
			}
			masterVocab.CheckAndReduceSimilarity()
			chSendBackVocab <- copyVocabWithNewSamples(masterVocab, vocab.Samples)
		case <-done:
			if shouldDedupe {
				dupes := findDupeSynapses(masterVocab.Net)
				for _, dupeGroup := range dupes {
					dedupeSynapses(dupeGroup, masterVocab.Net)
				}
			}
			return
		}
	}
}

/*
rerouteChangedIOCellIDs prevents nonexistent cells in the firing patterns of Inputs
and Outputs after merging to master vocab.

During a Network diff and merge, sometimes cells on the new network get new IDs.

When the cells with new IDs were part of the Input or Output firing pattern on the
branched networ/vocab, they would be referencing cells that do not exist on the
master network once they are merged to the master vocab.Inputs and vocab.Outputs.

This function updates those changed cell IDs on the vocab to be in sync with the
network.
*/
func rerouteChangedIOCellIDs(idChanges map[CellID]CellID, vocab *Vocabulary) {
	for _, outColl := range vocab.Outputs {
		for oldCellID, fires := range outColl.FirePattern {
			if newCellID, didChange := idChanges[oldCellID]; didChange {
				//log.Println("Swapping Output:", oldCellID, "for", newCellID)
				outColl.FirePattern[newCellID] = fires
				delete(outColl.FirePattern, oldCellID)
			}
		}
	}
	for _, vocabUnit := range vocab.Inputs {
		for oldCellID, fires := range vocabUnit.InputCells {
			if newCellID, didChange := idChanges[oldCellID]; didChange {
				//log.Println("Swapping Input:", oldCellID, "for", newCellID)
				vocabUnit.InputCells[newCellID] = fires
				delete(vocabUnit.InputCells, oldCellID)
			}
		}
	}
}
