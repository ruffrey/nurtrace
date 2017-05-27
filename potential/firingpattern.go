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
FireNetworkUntilDone takes some seed cells then fires the network until
it has no more firing, up to `laws.MaxPostFireSteps`.

Consider that you may want to ResetForTraining before running this.
*/
func FireNetworkUntilDone(network *Network, seedCells FiringPattern) (fp FiringPattern) {
	var i uint8
	fp = make(FiringPattern)
	for cellID := range seedCells {
		network.GetCell(cellID).FireActionPotential()
		network.resetCellsOnNextStep[cellID] = true
	}
	// we ignore the seedCells
	for {
		if i >= laws.MaxPostFireSteps {
			break
		}
		hasMore := network.Step()
		for cellID := range network.resetCellsOnNextStep {
			if _, ok := fp[cellID]; !ok {
				fp[cellID] = 0
			}
			fp[cellID]++
		}
		if !hasMore {
			break
		}
		i++
	}
	return fp
}

/*
FiringPatternDiff represents the firing differences between two
FiringPatterns.
*/
type FiringPatternDiff struct {
	Shared   map[CellID]uint16
	Unshared map[CellID]uint16
}

/*
Ratio is a measure of how alike the firing patterns of the diffed
cells were.
*/
func (diff *FiringPatternDiff) Ratio() float64 {
	lenShared := float64(len(diff.Shared))
	lenUnshared := float64(len(diff.Unshared))
	return lenShared / (lenShared + lenUnshared)
}

/*
mergeFiringPatterns returns a new FiringPattern containing a mashup of
the two supplied patterns.
*/
func mergeFiringPatterns(fp1, fp2 FiringPattern) (merged FiringPattern) {
	merged = make(FiringPattern)

	for cellID, fires := range fp1 {
		merged[cellID] = fires
	}
	for cellID, fires := range fp2 {
		if otherFires, already := merged[cellID]; already {
			merged[cellID] = uint16(math.Floor(float64(otherFires + fires/2)))
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
DiffFiringPatterns figures out what was alike and unshared between
two firing patterns.
*/
func DiffFiringPatterns(fp1, fp2 FiringPattern) *FiringPatternDiff {
	diff := &FiringPatternDiff{
		Shared:   make(map[CellID]uint16),
		Unshared: make(map[CellID]uint16),
	}

	for cellID, fires := range fp1 {
		if fp2Fires, ok := fp2[cellID]; ok {
			diff.Shared[cellID] = uint16(math.Abs(float64(fires) - float64(fp2Fires)))
		} else {
			diff.Unshared[cellID] = fires
		}
	}
	for cellID, fires := range fp2 {
		// already been through the shared ones
		if _, ok := fp1[cellID]; !ok {
			diff.Unshared[cellID] = fires
		}
	}
	return diff
}

/*
RunFiringPatternTraining trains the network using the training samples
in the vocab, until training samples are differentiated from one another.

The vocab should already be properly initiated and the network should be
set before running this.
*/
func RunFiringPatternTraining(vocab *Vocabulary, tag string) {
	vocab.Net.ResetForTraining()

	tots := len(vocab.Samples)
	fmt.Println(tag, "Running samples", tots)

	for iteration := 0; iteration < len(vocab.Samples); iteration++ {
		s := vocab.Samples[iteration]
		finalPattern := make(FiringPattern)

		if iteration%laws.TrainingResetIteration == 0 {
			fmt.Println(tag, "sample", iteration, "/", tots)
			vocab.Net.PrintTotals()
			fmt.Println(tag, "  outputs=", len(vocab.Outputs))
			vocab.Net.ResetForTraining()
		}
		// fire the input a bunch of times. after that we can consider
		// the output pattern as fired. set the output pattern.
		inputs := vocab.Inputs[s.input].InputCells
		for i := 0; i < laws.FiringIterationsPerSample; i++ {
			finalPattern = mergeFiringPatterns(finalPattern, FireNetworkUntilDone(vocab.Net, inputs))
		}
		if _, exists := vocab.Outputs[s.output]; !exists {
			vocab.Outputs[s.output] = NewOutputCollection(s.output)
		}
		// the output value is now represented by what we just
		// created above, merged with what we had before.
		originalFP := vocab.Outputs[s.output].FirePattern
		vocab.Outputs[s.output].FirePattern = mergeFiringPatterns(originalFP, finalPattern)

		// Now that the output firing pattern has been changed,
		// we need to ensure none of the other outputs are too similar.
		var lastOutput *OutputCollection
		atLeastOneWasTooSimilar := false
		i := -1
		for _, o := range vocab.Outputs {
			isThisOne := s.output == o.Value
			if isThisOne {
				continue
			}
			i++
			if i < 1 {
				lastOutput = o
				continue
			}
			fpDiff := DiffFiringPatterns(o.FirePattern, lastOutput.FirePattern)
			ratio := fpDiff.Ratio()
			// fmt.Println(tag, "Ratio:", lastOutput.Value, "vs", o.Value, "is", ratio)
			tooSimilar := ratio > laws.PatternSimilarityLimit
			if tooSimilar {
				atLeastOneWasTooSimilar = true
				// change input cell pattern
				expandInputs(vocab.Net, vocab.Inputs[s.input].InputCells)
				// change this output pattern
				expandOutputs(vocab.Net, ratio, o.FirePattern)
				// now re-run this one
				fmt.Println(tag, "RERUN:", lastOutput.Value, "vs", o.Value, "is", ratio)
			}
		}
		if atLeastOneWasTooSimilar {
			iteration--
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
	var closestRatio float64
	for _, outputCandidate := range vocab.Outputs {
		r := DiffFiringPatterns(outputCandidate.FirePattern, patt).Ratio()
		isCloser := r > closestRatio
		if isCloser {
			closestRatio = r
			oc = outputCandidate
		}
	}
	return oc
}
