package potential

import (
	"log"
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
cloneFiringPattern is necessary because maps are pass-by-reference,
and are often part of cloning
*/
func cloneFiringPattern(fp FiringPattern) FiringPattern {
	newFP := make(FiringPattern)
	for cellID, fires := range fp {
		newFP[cellID] = fires
	}
	return newFP
}

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
	network.FireNoise()
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
			combined := otherFires + fires
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
mergeAllInputs modifies the original group of Inputs by merging
every newer output collection into the original.
*/
func mergeAllInputs(original, newer map[InputValue]*VocabUnit) {
	for inVal, unit := range newer {
		if _, exists := original[inVal]; !exists {
			original[inVal] = unit
		} else {
			origFP := original[inVal].InputCells
			original[inVal].InputCells = mergeFiringPatterns(origFP, unit.InputCells)
		}
	}
}

/*
FiringPatternDiff represents the difference between two firing groups.
*/
type FiringPatternDiff struct {
	unshared map[CellID]float64
	shared   map[CellID]float64
	total    float64
}

/*
SimilarityRatio returns the ratio of alike to non-alike fires in a diff, and also
gives the unshared map as a firing pattern for easier use.
*/
func (fpdiff *FiringPatternDiff) SimilarityRatio() (similarity float64, unsharedCells FiringPattern) {
	allUnshared := 0.0
	unsharedCells = make(FiringPattern)
	for cellID, fires := range fpdiff.unshared {
		allUnshared += fires
		if fires < float64(math.MaxUint16) {
			unsharedCells[cellID] = uint16(fires)
		} else {
			unsharedCells[cellID] = math.MaxUint16
		}
	}
	for _, fires := range fpdiff.shared {
		allUnshared += fires
	}

	similarity = (fpdiff.total - allUnshared) / fpdiff.total
	return similarity, unsharedCells
}

/*
DiffFiringPatterns figures out what was alike and unshared between
two firing patterns.
*/
func DiffFiringPatterns(fp1, fp2 FiringPattern) *FiringPatternDiff {
	diff := FiringPatternDiff{
		unshared: make(map[CellID]float64),
		shared:   make(map[CellID]float64),
		total:    0.0,
	}

	for cellID, fires := range fp1 {
		fire1_64 := float64(fires)
		if fp2Fires, ok := fp2[cellID]; ok {
			fire2_64 := float64(fp2Fires)
			max := math.Max(fire1_64, fire2_64)
			min := math.Min(fire1_64, fire2_64)
			diff.total += max
			diff.shared[cellID] = max - min
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
func GetInputPatternForInputs(vocab *Vocabulary, inputs []InputValue) (cellsToFireForInputValues FiringPattern) {
	cellsToFireForInputValues = make(FiringPattern)
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
	tots := float64(len(vocab.Samples))
	errorred := 0.0

	for sampleIndex := 0; sampleIndex < len(vocab.Samples); sampleIndex++ {
		s := vocab.Samples[sampleIndex]
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
				expandInputs(vocab, cellsToFireForInputValues)
				continue
			}
			// just runs once if things fired
			break
		}

		originalFP := vocab.Outputs[s.output].FirePattern
		var newPattern FiringPattern
		closestOutput := FindClosestOutputCollection(sampleFirePattern, vocab)
		// Did this predict the right thing? If not, we just keep the sampleFirePattern
		// because the old pattern wasn't close enough.
		// TODO: is this a good rule?
		if closestOutput == nil {
			log.Println("sample:", s.inputs, "=", s.output, "; actual=nil", tag)
			errorred++
			// first timer, or not enough data, or not enough fired.
			newPattern = sampleFirePattern
		} else if closestOutput.Value == s.output {
			// predicted correctly
			log.Println("sample:", s.inputs, "=", s.output, "; correct", tag)
			newPattern = mergeFiringPatterns(originalFP, sampleFirePattern)
		} else {
			log.Println("sample:", s.inputs, "=", s.output,
				"; wrong=", closestOutput.Value, tag)
			errorred++
			// first timer, or poor prediction:
			// expected pattern gets overwritten
			expandInputs(vocab, cellsToFireForInputValues)
			newPattern = sampleFirePattern
			// actual pattern gets expanded for more uniqueness
			expandOutputs(vocab.Net, closestOutput.FirePattern, 1-laws.PatternSimilarityLimit)
		}

		vocab.Outputs[s.output].FirePattern = newPattern

		// sample is finished here, but provide an update on progress
		shouldRecalibrate := sampleIndex%laws.TrainingMergeBackIteration == 0
		if shouldRecalibrate {
			if sampleIndex != 0 { // not the first time
				chSynchVocab <- vocab
				vocab = <-chSendBackVocab
			}
			log.Println("progress", sampleIndex, "/", tots, tag)
		}
	}

	log.Println("Error=", errorred / tots, tag)
	chSynchVocab <- vocab
	vocab = <-chSendBackVocab
}

/*
CheckAndReduceSimilarity ensures none of the other outputs are too similar.

Each output pattern is compared to all other output patterns.

Should be called only on the master vocab, if doing multithreading.
*/
func (vocab *Vocabulary) CheckAndReduceSimilarity() {
	alreadyCompared := make(map[string]bool)
	totalOutputs := len(vocab.Outputs)
	outputCellMap := make(map[CellID]int)
	uselessCells := make(FiringPattern)

	//outputsUnique := make(map[OutputValue]map[CellID]bool)

	for _, primary := range vocab.Outputs {
		vocab.Net.ResetForTraining()
		for _, secondary := range vocab.Outputs {
			isThisOne := secondary.Value == primary.Value
			if isThisOne {
				continue
			}
			// only run each comparison once
			var key string
			sPri := string(primary.Value)
			sSec := string(secondary.Value)
			if primary.Value > secondary.Value {
				key = sPri + sSec
			} else {
				key = sSec + sPri
			}
			if alreadyCompared[key] {
				continue
			}
			alreadyCompared[key] = true

			// cause dissimiliarity between the unshared cells in the
			// two patterns
			diff := DiffFiringPatterns(primary.FirePattern, secondary.FirePattern)
			ratio, unsharedFiringPattern := diff.SimilarityRatio()
			tooSimilar := ratio > laws.PatternSimilarityLimit
			if tooSimilar {
				// change this output pattern
				//log.Println("EXPAND:", secondary.Value, "vs", primary.Value, "is", ratio)
				expandOutputs(vocab.Net, unsharedFiringPattern, ratio)
			}
		}
		// track cells in map so we can see which cells don't
		// provide new information later
		for cellID := range primary.FirePattern {
			if _, exists := outputCellMap[cellID]; !exists {
				outputCellMap[cellID] = 0
			}
			outputCellMap[cellID]++
		}
	}

	// which cells are fired on every expected output?
	for cellID, outputsThatHaveIt := range outputCellMap {
		if outputsThatHaveIt >= totalOutputs {
			uselessCells[cellID] = 1
		}
	}
	log.Println("Noise=", float64(len(uselessCells))/float64(len(vocab.Net.Cells)))
	// turned off because it isn't clear whether this really helps or not
	vocab.Noise = uselessCells
}

/*
FindClosestOutputCollection finds the closest output collection
by simple average.

patt = the actual firing pattern

This is useful for sampling.

Subtracting noise is done.
*/
func FindClosestOutputCollection(patt FiringPattern, vocab *Vocabulary) (oc *OutputCollection) {
	closestRatio := 0.0
	slimPatt := removeNoise(vocab.Noise, patt)
	for _, outputCandidate := range vocab.Outputs {
		noiselessPattern := removeNoise(vocab.Noise, outputCandidate.FirePattern)
		r, _ := DiffFiringPatterns(slimPatt, noiselessPattern).SimilarityRatio()
		isCloser := r > closestRatio
		if isCloser {
			closestRatio = r
			oc = outputCandidate
		}
	}
	return oc
}

// removeNoise returns a new firing pattern with noise removed
func removeNoise(noisePatt, fp FiringPattern) (denoisedFP FiringPattern) {
	denoisedFP = cloneFiringPattern(fp)
	for cellID, noiseReductionFires := range noisePatt {
		if fireTotal, hasNoise := denoisedFP[cellID]; hasNoise {
			noiselessValue := fireTotal - noiseReductionFires
			if noiselessValue < 0 {
				noiselessValue = 0
			}
			denoisedFP[cellID] = noiselessValue
		}
	}
	return denoisedFP
}
