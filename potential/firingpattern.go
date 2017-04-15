package potential

import (
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
		network.getCell(cellID).FireActionPotential()
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
DifferentiateVocabUnits grows two firing groups until they produce significantly
different patterns from one another.

Does not modify the network.
*/
func DifferentiateVocabUnits(vu1, vu2 *VocabUnit, _network *Network) Diff {
	// add general noise
	n := CloneNetwork(_network)

	for {
		n.ResetForTraining()
		vu1.ExpandExistingInputs(n)
		for i := 0; i < laws.FiringIterationsPerSample-1; i++ {
			FireNetworkUntilDone(n, vu1.InputCells)
		}
		patt1 := FireNetworkUntilDone(n, vu1.InputCells)
		n.ResetForTraining()
		vu2.ExpandExistingInputs(n)
		for i := 0; i < laws.FiringIterationsPerSample-1; i++ {
			FireNetworkUntilDone(n, vu2.InputCells)
		}
		patt2 := FireNetworkUntilDone(n, vu2.InputCells)
		n.ResetForTraining()

		fpDiff := DiffFiringPatterns(patt1, patt2)
		if fpDiff.Ratio() < laws.PatternSimilarityLimit {
			return DiffNetworks(_network, n)
		}
	}
}

/*
RunFiringPatternTraining trains the network using the training samples,
until training samples are differentiated from one another.
*/
func RunFiringPatternTraining(network *Network, vocab []VocabUnit) {

}

/*
FindClosestOutputCollection finds the closest output collection
statisitcally.
*/
func FindClosestOutputCollection(patt *FiringPattern, vocab *Vocabulary) (ov OutputValue) {

	return ov
}
