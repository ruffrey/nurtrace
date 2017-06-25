package potential

import (
	"math"

	"github.com/ruffrey/nurtrace/laws"
)

// InputValue is a unique string for the input
type InputValue string

/*
A VocabUnit is a group of cells that represent a user-defined value.

VocabUnit is similar to the previous PerceptionUnit, when cells
were single input and output. Here, a group of cells are the input.
*/
type VocabUnit struct {
	Value      InputValue
	InputCells FiringPattern
}

/*
NewVocabUnit is a factory for VocabUnit
*/
func NewVocabUnit(value string) *VocabUnit {
	return &VocabUnit{
		Value:      InputValue(value),
		InputCells: make(FiringPattern),
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
func NewOutputCollection(value OutputValue) *OutputCollection {
	return &OutputCollection{
		Value:       OutputValue(value),
		FirePattern: make(FiringPattern),
	}
}

/*
InitRandomInputs chooses some input cells for the vocab unit.
*/
func (vu *VocabUnit) InitRandomInputs(vocab *Vocabulary) {
	for i := 0; i < laws.InitialCellCountPerVocabUnit; i++ {
		var inputCellID CellID
		var depthGrowerCellID CellID // attempt to ensure there are deep paths in the net
		for {
			inputCellID = vocab.Net.RandomCellKey()
			if !isCellOnAnyInput(inputCellID, vocab.Inputs) {
				break
			}
		}
		for {
			depthGrowerCellID = vocab.Net.RandomCellKey()
			if !isCellOnAnyInput(depthGrowerCellID, vocab.Inputs) {
				break
			}
		}
		vu.InputCells[inputCellID] = 1
		vocab.Net.GrowPathBetween(inputCellID, depthGrowerCellID, laws.ComputedSynapsesPerCell)
	}
}

func isCellOnAnyInput(cellID CellID, inputs map[InputValue]*VocabUnit) bool {
	for _, vu := range inputs {
		if _, exists := vu.InputCells[cellID]; exists {
			return true
		}
	}
	return false
}

func randCellFromFP(cellMap FiringPattern) (randCellID CellID) {
	iterate := randomIntBetween(0, len(cellMap)-1)
	i := 0
	for k := range cellMap {
		if i == iterate {
			randCellID = CellID(k)
			break
		}
		i++
	}
	return randCellID
}

/*
expandInputs expands an input firing pattern by a set number of synapses.
*/
func expandInputs(vocab *Vocabulary, fp FiringPattern) {
	for i := 0; i < laws.InputCellDifferentiationCount; i++ {
		preCell := randCellFromFP(fp)
		// Do not just fire another input cell; that would
		// be a little confounding right out of the gate.
		for {
			anotherCell := vocab.Net.RandomCellKey()
			if !isCellOnAnyInput(anotherCell, vocab.Inputs) {
				vocab.Net.GrowPathBetween(preCell, anotherCell, laws.ComputedSynapsesPerCell)
				break
			}
		}

	}
}

// Not greater than 10% more synapses than the ideal number of synapses per cell.
const _maxSynapseRatio float64 = laws.IdealCellSynapseBalance - (laws.IdealCellSynapseBalance * .1)
/*
expandOutputs expands an output firing pattern by adding more
synapses from the firting pattern to random cells.
*/
func expandOutputs(network *Network, unsharedCellsFP FiringPattern, similarity float64) {
	totalUnshared := float64(len(unsharedCellsFP))
	pctOverLimit := similarity - laws.PatternSimilarityLimit
	uniquenessToAdd := int(math.Ceil(pctOverLimit * totalUnshared))

	currentRatio := float64(len(network.Cells)) / float64(len(network.Synapses))
	shouldAddCell := currentRatio > laws.IdealCellSynapseBalance || currentRatio < _maxSynapseRatio

	for i := 0; i < uniquenessToAdd; i++ {
		preCell := randCellFromFP(unsharedCellsFP)
		var nextCell CellID

		if shouldAddCell {
			nextCell = NewCell(network).ID
			network.linkCells(nextCell, network.RandomCellKey())
		} else {
			nextCell = network.RandomCellKey()
		}

		network.GrowPathBetween(preCell, nextCell, laws.ComputedSynapsesPerCell)
	}
}
