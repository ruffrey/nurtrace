package charrnn

import (
	"bleh/potential"
	"strings"
)

/*
VocabItem has a cell for the input and the output, both of which represent the same word.

It would have been just as easy to have separate inputs and outputs.
*/
type VocabItem struct {
	Character  string
	InputCell  potential.CellID
	OutputCell potential.CellID
}

/*
Vocab represents the vocab
*/
type Vocab struct {
	CharToItem map[string]VocabItem
	CellToChar map[potential.CellID]string
}

/*
NewVocab generates the vocab from a bunch of text
*/
func NewVocab(text string, network *potential.Network) Vocab {
	chars := strings.Split(text, "")
	vocab := Vocab{
		CharToItem: make(map[string]VocabItem),
		CellToChar: make(map[potential.CellID]string),
	}

	for _, Character := range chars {
		if _, exists := vocab.CharToItem[Character]; !exists {
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
			network.Cells[InputCell].Tag = "in-" + Character
			network.Cells[OutputCell].Tag = "out-" + Character

			network.GrowPathBetween(InputCell, OutputCell, 10)

			vocab.CharToItem[Character] = VocabItem{
				Character,
				InputCell,
				OutputCell,
			}
		}
	}

	start := VocabItem{
		Character: "START",
		InputCell: network.RandomCellKey(),
	}
	for {
		start.OutputCell = network.RandomCellKey()
		if start.InputCell != start.OutputCell {
			break
		}
	}
	end := VocabItem{
		Character: "END",
		InputCell: network.RandomCellKey(),
	}
	for {
		end.OutputCell = network.RandomCellKey()
		if end.InputCell != end.OutputCell {
			break
		}
	}

	network.GrowPathBetween(start.InputCell, start.OutputCell, 10)
	network.GrowPathBetween(end.InputCell, end.OutputCell, 10)

	network.Cells[start.InputCell].Tag = "in-START"
	network.Cells[start.OutputCell].Tag = "out-START"
	network.Cells[end.InputCell].Tag = "in-END"
	network.Cells[end.OutputCell].Tag = "out-END"

	vocab.CharToItem["START"] = start
	vocab.CharToItem["END"] = end

	for char, vi := range vocab.CharToItem {
		vocab.CellToChar[vi.InputCell] = char
		vocab.CellToChar[vi.OutputCell] = char
		network.Cells[vi.InputCell].Immortal = true
		network.Cells[vi.OutputCell].Immortal = true
	}

	return vocab
}
