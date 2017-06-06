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
		network.Cells[cellID].activating = true
	}
	// we ignore the seedCells
	for {
		if i >= laws.MaxPostFireSteps {
			break
		}
		hasMore := network.Step()
		for _cellID, cell := range network.Cells {
			cellID := CellID(_cellID)
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
DiffFiringPatterns figures out what was alike and unshared between
two firing patterns.
*/
func DiffFiringPatterns(fp1, fp2 FiringPattern) float64 {
	unshared := make(map[CellID]float64)

	total := 0.0
	for cellID, fires := range fp1 {
		fire1_64 := float64(fires)
		if fp2Fires, ok := fp2[cellID]; ok {
			fire2_64 := float64(fp2Fires)
			max := math.Max(fire1_64, fire2_64)
			min := math.Min(fire1_64, fire2_64)
			total += max
			unshared[cellID] = max - min
		} else {
			unshared[cellID] = fire1_64
			total += fire1_64
		}
	}
	for cellID, fires := range fp2 {
		fire1_64 := float64(fires)
		// already been through the shared ones
		if _, ok := fp1[cellID]; !ok {
			unshared[cellID] = fire1_64
			total += fire1_64
		}
	}

	allUnshared := 0.0
	for _, fires := range unshared {
		allUnshared += fires
	}
	return (total - allUnshared) / total
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
			ratio := DiffFiringPatterns(o.FirePattern, lastOutput.FirePattern)
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
		r := DiffFiringPatterns(outputCandidate.FirePattern, patt)
		isCloser := r > closestRatio
		if isCloser {
			closestRatio = r
			oc = outputCandidate
		}
	}
	return oc
}
