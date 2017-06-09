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
func (vu *VocabUnit) InitRandomInputs(network *Network) {
	for i := 0; i < laws.InitialCellCountPerVocabUnit; i++ {
		vu.InputCells[network.RandomCellKey()] = 1
	}
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
		preCell2 := randCellFromFP(fp)
		newCell := NewCell(network)

		newSynapse1 := network.linkCells(preCell1, newCell.ID)
		newSynapse2 := network.linkCells(preCell2, newCell.ID)
		// excitatory
		newSynapse1.Millivolts = int16(math.Abs(float64(newSynapse1.Millivolts)))
		newSynapse2.Millivolts = int16(math.Abs(float64(newSynapse2.Millivolts)))
		// fp[newCell.ID] = true
	}
}

/*
expandOutputs expands an output firing pattern by adding more INHIBITORY
synapses to new cells that are attached to random cells in its map.
*/
func expandOutputs(network *Network, noisyCellsToAdd int, fp FiringPattern) {
	for i := 0; i < noisyCellsToAdd; i++ {
		preCell := randCellFromFP(fp)
		newCell := NewCell(network)
		newSynapse := network.linkCells(preCell, newCell.ID)
		// inhibitory
		newSynapse.Millivolts = -int16(math.Abs(float64(newSynapse.Millivolts)))
	}
}
