package potential

import (
	"fmt"

	"github.com/ruffrey/nurtrace/laws"
)

/*
Glossary:
- VocabUnit - single input value mapped to multiple input cells
- OutputCollection - single output value mapped to multiple output cells
*/

// FiringPattern represents all cells that fired in a single step
type FiringPattern map[CellID]bool

/*
FireNetworkUntilDone takes some seed cells then fires the network until
it has no more firing, up to `laws.MaxPostFireSteps`.

Consider that you may want to ResetForTraining before running this.
*/
func FireNetworkUntilDone(network *Network, seedCells map[CellID]bool) (fp FiringPattern) {
	var i uint8
	fp = make(map[CellID]bool)
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
			fp[cellID] = true
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
	Shared   map[CellID]bool
	Unshared map[CellID]bool
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

	for cellID := range fp1 {
		merged[cellID] = true
	}
	for cellID := range fp2 {
		merged[cellID] = true
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
		Shared:   make(map[CellID]bool),
		Unshared: make(map[CellID]bool),
	}

	for cellID := range fp1 {
		if fp2[cellID] {
			diff.Shared[cellID] = true
		} else {
			diff.Unshared[cellID] = true
		}
	}
	for cellID := range fp2 {
		// already been through the shared ones
		if !fp1[cellID] {
			diff.Unshared[cellID] = true
		}
	}
	return diff
}

/*
RunFiringPatternTraining trains the network using the training samples
in the vocab, until training samples are differentiated from one another.

The vocab should already be properly initiated and the network should be
set before running this.

TODO: make multithreaded and multi-workered
*/
func RunFiringPatternTraining(vocab *Vocabulary) {
	vocab.Net.ResetForTraining()
	finalPattern := make(FiringPattern)

	tots := len(vocab.Samples)
	fmt.Println("Running samples", tots)

	for iteration := 0; iteration < len(vocab.Samples); iteration++ {
		s := vocab.Samples[iteration]
		if iteration%laws.TrainingResetIteration == 0 {
			fmt.Println("sample", iteration, "/", tots)
			vocab.Net.PrintTotals()
			fmt.Println("  outputs=", len(vocab.Outputs))
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
			// fmt.Println("Ratio:", lastOutput.Value, "vs", o.Value, "is", ratio)
			tooSimilar := ratio > laws.PatternSimilarityLimit
			// TODO: wrong because we decrement iteration which is
			// from ABOVE!
			if tooSimilar {
				expandFiringPattern(vocab.Net, lastOutput.FirePattern)
				expandFiringPattern(vocab.Net, o.FirePattern)
				// now re-run this one
				fmt.Println("RERUN:", lastOutput.Value, "vs", o.Value, "is", ratio)
				iteration--
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
