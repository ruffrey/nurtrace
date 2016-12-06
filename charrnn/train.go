package charrnn

import (
	"bleh/potential"
	"strings"
)

// TODO: this file and charrnn.go will be merged

/*
Charrnn is
*/
type Charrnn struct {
	potential.Trainer
	potential.TrainingSettings
}

/*
PrepareData is from potential.Trainer
*/
func (charrnn *Charrnn) PrepareData(network *potential.Network) {
	chars := strings.Split(text, "")
	charrnn.Data = Vocab{
		KeyToItem: make(map[string]VocabItem),
		CellToKey: make(map[potential.CellID]string),
	}

	for _, Value := range chars {
		if _, exists := charrnn.Data.KeyToItem[Value]; !exists {
			InputCell := network.RandomCellKey()
			// ensure the input and output cells are not the same!
			var OutputCell potential.CellID
			for {
				OutputCell = network.RandomCellKey()
				if OutputCell != InputCell {
					break
				}
			}
			// inputs and outputs must never be pruned
			network.Cells[InputCell].Tag = "in-" + Value
			network.Cells[OutputCell].Tag = "out-" + Value

			charrnn.Data.KeyToItem[Value] = VocabItem{
				Value,
				InputCell,
				OutputCell,
			}
		}
	}

	start := VocabItem{
		Value:     "START",
		InputCell: network.RandomCellKey(),
	}
	for {
		start.OutputCell = network.RandomCellKey()
		if start.InputCell != start.OutputCell {
			break
		}
	}
	end := VocabItem{
		Value:     "END",
		InputCell: network.RandomCellKey(),
	}
	for {
		end.OutputCell = network.RandomCellKey()
		if end.InputCell != end.OutputCell {
			break
		}
	}

	network.Cells[start.InputCell].Tag = "in-START"
	network.Cells[start.OutputCell].Tag = "out-START"
	network.Cells[end.InputCell].Tag = "in-END"
	network.Cells[end.OutputCell].Tag = "out-END"

	charrnn.Data.KeyToItem["START"] = start
	charrnn.Data.KeyToItem["END"] = end

	return charrnn.Data
}
