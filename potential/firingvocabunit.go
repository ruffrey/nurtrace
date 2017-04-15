package potential

import "github.com/ruffrey/nurtrace/laws"

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
