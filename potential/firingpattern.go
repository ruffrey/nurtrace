package potential

import (
	"fmt"
	"math"

	"github.com/ruffrey/nurtrace/laws"
)

/*
Glossary:
- VocabUnit - single input value mapped to multiple input cells
- OutputCollection - single output value mapped to multiple output cells
*/

// FiringPattern represents all cells that fired in a single step
type FiringPattern map[CellID]uint16

/*
FireNetworkUntilDone takes some seed cells, fires them,
then fires the network until it has no more firing - up
to `laws.MaxPostFireSteps`.

Consider that you may want to ResetForTraining before running this.
*/
func FireNetworkUntilDone(network *Network, seedCells FiringPattern) FiringPattern {
	i := 0
	fp := make(FiringPattern)

	// Do the seeding
	for ; i < laws.FiringIterationsPerSample; i++ {
		for cellID := range seedCells {
			network.GetCell(cellID).FireActionPotential()
		}
		network.Step()
	}

	i = 0
	for {
		if i >= laws.MaxPostFireSteps {
			break
		}

		hasMore := network.Step()
		for _cellID, cell := range network.Cells {
			cellID := CellID(_cellID)
			// skip seed cells on first round because they will
			// still be activating
			if i == 0 {
				if _, isSeedCell := seedCells[cellID]; isSeedCell {
					continue
				}
			}
			if cell.activating {
				if _, ok := fp[cellID]; !ok {
					fp[cellID] = 0
				}
				fp[cellID]++
			}
		}

		if !hasMore {
			break
		}

		i++
	}
	return fp
}

/*
mergeFiringPatterns returns a new FiringPattern containing an average of
the two supplied patterns.
*/
func mergeFiringPatterns(fp1, fp2 FiringPattern) (merged FiringPattern) {
	merged = make(FiringPattern)

	for cellID, fires := range fp1 {
		merged[cellID] = fires
	}
	for cellID, fires := range fp2 {
		if otherFires, already := merged[cellID]; already {
			combined := (otherFires + fires)
			average := float64(combined / 2)
			bottomValue := uint16(math.Floor(average))
			merged[cellID] = bottomValue
		} else {
			merged[cellID] = fires
		}
	}

	return merged
}

/*
mergeAllOutputs modifies the original group of Outputs by merging
every newer output collection into the original.
*/
func mergeAllOutputs(original, newer map[OutputValue]*OutputCollection) {
	for outVal, collection := range newer {
		if _, exists := original[outVal]; !exists {
			original[outVal] = collection
		} else {
			origFP := original[outVal].FirePattern
			original[outVal].FirePattern = mergeFiringPatterns(origFP, collection.FirePattern)
		}
	}
}

/*
FiringPatternDiff represents the difference between two firing groups.
*/
type FiringPatternDiff struct {
	unshared map[CellID]float64
	shared   FiringPattern
	total    float64
}

/*
SimilarityRatio returns the ratio of alike to non-alike fires in a diff, and also
gives the unshared map as a firing pattern for easier use.
*/
func (fpdiff *FiringPatternDiff) SimilarityRatio() (float64, FiringPattern) {
	allUnshared := 0.0
	fp := make(FiringPattern)
	for cellID, fires := range fpdiff.unshared {
		allUnshared += fires
		if fires < float64(math.MaxUint16) {
			fp[cellID] = uint16(fires)
		} else {
			fp[cellID] = math.MaxUint16
		}
	}
	return (fpdiff.total - allUnshared) / fpdiff.total, fp
}

/*
DiffFiringPatterns figures out what was alike and unshared between
two firing patterns.
*/
func DiffFiringPatterns(fp1, fp2 FiringPattern) *FiringPatternDiff {
	diff := FiringPatternDiff{
		unshared: make(map[CellID]float64),
		shared:   make(FiringPattern),
		total:    0.0,
	}

	for cellID, fires := range fp1 {
		fire1_64 := float64(fires)
		if fp2Fires, ok := fp2[cellID]; ok {
			fire2_64 := float64(fp2Fires)
			max := math.Max(fire1_64, fire2_64)
			min := math.Min(fire1_64, fire2_64)
			diff.total += max
			diff.unshared[cellID] = max - min
		} else {
			diff.unshared[cellID] = fire1_64
			diff.total += fire1_64
		}
	}
	for cellID, fires := range fp2 {
		fire1_64 := float64(fires)
		// already been through the shared ones
		if _, ok := fp1[cellID]; !ok {
			diff.unshared[cellID] = fire1_64
			diff.total += fire1_64
		}
	}

	return &diff
}

/*
GetInputPatternForInputs accepts an array of input characters and
returns a merged firing pattern to be fired for that entire group
of inputs and their underlying cells.

For example, given the inputs `[]{"a", "b", "c"}` it will return the
input cells to be fired for these three input values.
*/
func GetInputPatternForInputs(vocab *Vocabulary, inputs []InputValue) FiringPattern {
	cellsToFireForInputValues := make(FiringPattern)
	for _, inputChar := range inputs {
		cellsForInputChar := vocab.Inputs[inputChar]

		cellsToFireForInputValues = mergeFiringPatterns(
			cellsToFireForInputValues, cellsForInputChar.InputCells)
	}

	return cellsToFireForInputValues
}

/*
RunFiringPatternTraining trains the network using the training samples
in the vocab, until training samples are differentiated from one another.

The vocab should already be properly initiated and the network should be
set before running this.
*/
func RunFiringPatternTraining(vocab *Vocabulary, chSynchVocab chan *Vocabulary, chSendBackVocab chan *Vocabulary, tag string) {
	tots := len(vocab.Samples)
	fmt.Println(tag, "Running samples", tots)

	for sampleIndex, s := range vocab.Samples {
		var sampleFirePattern FiringPattern

		// merge the inputs first
		cellsToFireForInputValues := GetInputPatternForInputs(vocab, s.inputs)

		vocab.Net.ResetForTraining()

		// Training

		// fire the inputs a bunch of times. after that we can consider
		// the output pattern as fired. set the output pattern.
		for {
			sampleFirePattern = FireNetworkUntilDone(vocab.Net, cellsToFireForInputValues)
			nothingFired := len(sampleFirePattern) == 0
			if nothingFired {
				expandInputs(vocab.Net, cellsToFireForInputValues)
				continue
			}
			// just runs once if things fired
			break
		}

		// the output value is now represented by what we just
		// created above, merged with what we had before.
		originalFP := vocab.Outputs[s.output].FirePattern
		vocab.Outputs[s.output].FirePattern = mergeFiringPatterns(originalFP, sampleFirePattern)

		// sample is finished here, but provide an update on progress
		shouldRecalibrate := sampleIndex%laws.TrainingMergeBackIteration == 0
		if shouldRecalibrate {
			if sampleIndex != 0 { // not the first time
				chSynchVocab <- vocab
				vocab = <-chSendBackVocab
			}
			fmt.Println(tag, "progress", sampleIndex, "/", tots)
		}
	}

	chSynchVocab <- vocab
	vocab = <-chSendBackVocab
}

/*
CheckAndReduceSimilarity ensures none of the other outputs are too similar.

Each output pattern is compared to all other output patterns.

Argument `tag` is for logging.
*/
func (vocab *Vocabulary) CheckAndReduceSimilarity() {
	alreadyCompared := make(map[string]bool)
	totalOutputs := len(vocab.Outputs)
	outputCellMap := make(map[CellID]int)
	uselessCells := make(map[CellID]bool)

	// returns whether the outputs were too similar, and we need to recheck
	// the output collection. This has the effect of making sure outputs are
	// sufficiently different from one another, before moving on
	runMap := func(primary *OutputCollection, secondary *OutputCollection) bool {
		isThisOne := secondary.Value == primary.Value
		if isThisOne {
			return false
		}
		var key string
		sPri := string(primary.Value)
		sSec := string(secondary.Value)
		if primary.Value > secondary.Value {
			key = sPri + sSec
		} else {
			key = sSec + sPri
		}
		if alreadyCompared[key] {
			return false
		}
		alreadyCompared[key] = true

		diff := DiffFiringPatterns(primary.FirePattern, secondary.FirePattern)
		ratio, unsharedFiringPattern := diff.SimilarityRatio()
		tooSimilar := ratio > laws.PatternSimilarityLimit
		if tooSimilar {
			// change this output pattern
			expandOutputs(vocab.Net, unsharedFiringPattern)
			 //fmt.Println(tag, "EXPAND:", secondary.Value, "vs", primary.Value, "is", ratio)
			return true
		}
		return false
	}

	for _, primary := range vocab.Outputs {
		vocab.Net.ResetForTraining()
		for _, secondary := range vocab.Outputs {
			for {
				rerun := runMap(primary, secondary)
				if !rerun {
					break
				}
			}
		}
		// track cells in map so we can see which cells don't
		// provide new information
		for cellID := range primary.FirePattern {
			if _, exists := outputCellMap[cellID]; !exists {
				outputCellMap[cellID] = 0
			}
			outputCellMap[cellID]++
		}
	}

	for cellID, outputsThatHaveIt := range outputCellMap {
		if outputsThatHaveIt >= totalOutputs {
			uselessCells[cellID] = true
		}
	}
	totalUseless := len(uselessCells)
	if totalUseless > 0 {
		fmt.Println("Useless:\n ", totalUseless)
		for _, oc := range vocab.Outputs {
			for cellID := range uselessCells {
				delete(oc.FirePattern, cellID)
			}
		}
	}
}

/*
FindClosestOutputCollection finds the closest output collection
statisitcally.

patt = the actual firing pattern

This is useful for sampling.
*/
func FindClosestOutputCollection(patt FiringPattern, vocab *Vocabulary) (oc *OutputCollection) {
	closestRatio := 0.0
	for _, outputCandidate := range vocab.Outputs {
		r, _ := DiffFiringPatterns(patt, outputCandidate.FirePattern).SimilarityRatio()
		isCloser := r > closestRatio
		if isCloser {
			closestRatio = r
			oc = outputCandidate
		}
	}
	return oc
}
