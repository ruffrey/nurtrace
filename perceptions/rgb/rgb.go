package rgb

import (
	"strings"

	"github.com/ruffrey/nurtrace/perceptions/charrnn"
)

/*
Charcat is a neural network model that goes from groups of characters
to a predefined category. It uses the charrnn package. But instead of
characters predicting characters, the characters predict a category.
*/
type Charcat struct {
	rawText string
	charrnn.Charrnn
}

// type rgbPerceptionUnit struct {
// 	Value      []uint8
// 	InputCell  potential.CellID
// 	OutputCell potential.CellID
// 	RGB        []uint8
// 	Name       string
// 	Hex        string
// 	OutputCell potential.CellID
// }

// SetRawData sets the rawText prop
func (charcat *Charcat) SetRawData(bytes []byte) {
	text := string(bytes)
	charcat.rawText = text
	charcat.Chars = strings.Split(text, "")
}
