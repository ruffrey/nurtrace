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
		var cellID CellID
		for {
			cellID = vocab.Net.RandomCellKey()
			if !_isCellOnAnyInput(cellID, vocab.Inputs) {
				break
			}
		}
		vu.InputCells[cellID] = 1
	}
}

func _isCellOnAnyInput(cellID CellID, inputs map[InputValue]*VocabUnit) bool {
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
expandInputs expands an input firing pattern by a set number of cells and uses
an excitatory synapse to link them.
*/
func expandInputs(network *Network, fp FiringPattern) {
	for i := 0; i < laws.InputCellDifferentiationCount; i++ {
		preCell1 := randCellFromFP(fp)
		newCell := NewCell(network)

		newSynapse1 := network.linkCells(preCell1, newCell.ID)
		// excitatory
		newSynapse1.Millivolts = int16(math.Abs(float64(newSynapse1.Millivolts)))
	}
}

/*
expandOutputs expands an output firing pattern by adding more
synapses from the firting pattern to random cells.
*/
func expandOutputs(network *Network, unsharedCellsFP FiringPattern, similarity float64) {
	totalUnshared := float64(len(unsharedCellsFP))
	pctOverLimit := similarity - laws.PatternSimilarityLimit
	uniquenessToAdd := int(math.Ceil(pctOverLimit * totalUnshared))
	for i := 0; i < uniquenessToAdd; i++ {
		var preCell CellID
		// fewer than this means we need to add more unique cells
		if totalUnshared > laws.MinUniqueCellsDuringExpand {
			preCell = randCellFromFP(unsharedCellsFP)
		} else {
			preCell = NewCell(network).ID
		}
		network.linkCells(preCell, network.RandomCellKey())

		// also reinforce a random unshared synapse
		anotherCell := network.GetCell(randCellFromFP(unsharedCellsFP))
		if len(anotherCell.AxonSynapses) != 0 {
			randReinforceSynapse := randSynapseFromMap(anotherCell.AxonSynapses)

			network.GetSyn(randReinforceSynapse).reinforce()
		}
	}
}
