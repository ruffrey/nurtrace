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
Vocab represnts the vocab
*/
type Vocab map[string]VocabItem

/*
NewVocab generates the vocab from a bunch of text
*/
func NewVocab(text string, network *potential.Network) Vocab {
	chars := strings.Split(text, "")
	vocab := make(Vocab)

	for _, Character := range chars {
		if _, exists := vocab[Character]; !exists {
			InputCell := network.RandomCellKey()
			// ensure the input and output cells are not the same!
			var OutputCell potential.CellID
			for {
				OutputCell = network.RandomCellKey()
				if OutputCell != InputCell {
					break
				}
			}
			vocab[Character] = VocabItem{
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
	vocab["START"] = start
	vocab["END"] = end

	return vocab
}
