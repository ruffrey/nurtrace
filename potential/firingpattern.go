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

// InputValue is a unique string for the input
type InputValue string

/*
A VocabUnit is a group of cells that represent a user-defined value.

VocabUnit is similar to the previous PerceptionUnit, when cells
were single input and output. Here, a group of cells are the input.
*/
type VocabUnit struct {
	Value      InputValue
	InputCells map[CellID]bool
}

/*
NewVocabUnit is a factory for VocabUnit
*/
func NewVocabUnit(value string) *VocabUnit {
	return &VocabUnit{
		Value:      InputValue(value),
		InputCells: make(map[CellID]bool),
	}
}

// OutputValue is a unique string for an output, useful for identifying it.
type OutputValue string

/*
OutputCollection is a firing pattern that represents an output value. It is
a collection of cells, which can change. These cells represent a value.
*/
type OutputCollection struct {
	Value       OutputValue
	FirePattern FiringPattern
}

/*
NewOutputCollection is a factory for OutputCollection
*/
func NewOutputCollection(value string) *OutputCollection {
	return &OutputCollection{
		Value: OutputValue(value),
	}
}

/*
InitRandomInputs chooses some input cells for the vocab unit.
*/
func (vu *VocabUnit) InitRandomInputs(network *Network) {
	for i := 0; i < laws.InitialCellCountPerVocabUnit; i++ {
		vu.InputCells[network.RandomCellKey()] = true
	}
}

/*
ExpandExistingInputs grows out from the vocab unit's existing InputCells
so it has more uniqueness.
*/
func (vu *VocabUnit) ExpandExistingInputs(network *Network) {
	for i := 0; i < laws.NewCellDifferentiationCount; i++ {
		preCell := randCellFromMap(vu.InputCells)
		network.GrowPathBetween(preCell, NewCell(network).ID, laws.GrowPathExpectedMinimumSynapses)
	}
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
