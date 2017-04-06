package charcat

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ruffrey/nurtrace/potential"
)

type inputTextChars string
type outputCategoryName string

/*
Charcatnn predicts a category from strings of characters.

Test data should be structured as JSON:

	[
		{ "CategoryName": "gray", "InputText": "cccccc" }
	]

Category names *must* be greater than one character!

A character-category neural network will have each of the cells from
the input text fired. So in the case of `cccccc`, it would fire
the "c" cell six times in a row. We would predict that the category
cell for "gray" would fire, and no other category cells should fire.
In reality, gray, gray23, Light Gray all might fire.

One way this differes from a Charrnn is we do not track the START and
END. There will always be a specified number of chars in input
text which yields a single category. Versus with Charrnn, text streams
from START and produces output until we hit END.
*/
type Charcatnn struct {
	structuredTrainingCases []charcatTrainingData
	// settings *potential.TrainingSettings
	// perception.Perception
}

type charcatTrainingData struct {
	CategoryName string // `dark gray`
	InputText    string // `1e1e1e`
}

/*
charcatPerceptionUnit is the format of this model's *Vocab*, with mappings to the
cell which represents it. A perception unit can represent either the category name
or the input text; not both.
*/
type charcatPerceptionUnit struct {
	CategoryName string
	InputChar    string
	CellID       potential.CellID
}

/*
SaveVocab saves the current vocabulary from the Charcat network.

The key is the input chars (hex color key) *or* output category name. Since vocab
represents a mapping from either kind of data (input or output) back to a cell.

	[
		{ "CategoryName": "gray", CellID: 3577432 }, // output
		{ "InputChar": "c", CellID: 588433 },	     // input
		{ "InputChar": "0", CellID: 333423 },	     // input
	]

*/
func (charcat *Charcatnn) SaveVocab(settings *potential.TrainingSettings, filename string) error {
	var data []charcatPerceptionUnit
	for key, genericPercepUnit := range settings.Data.KeyToItem {
		sKey := key.(string)
		var percepUnit charcatPerceptionUnit
		// fmt.Println("Preparing save perceptionUnit", key, genericPercepUnit)
		if genericPercepUnit.OutputCell != 0 {
			percepUnit.CategoryName = sKey
			percepUnit.CellID = genericPercepUnit.OutputCell

		} else {
			percepUnit.InputChar = sKey
			percepUnit.CellID = genericPercepUnit.InputCell
		}

		if percepUnit.CategoryName == "" && percepUnit.InputChar == "" {
			fmt.Println(percepUnit)
			fmt.Println(sKey, genericPercepUnit)
			return errors.New("Bad perception unit in Data.KeyToItem: unable to map between cell and category or input character")
		}
		if percepUnit.CellID == 0 {
			fmt.Println(percepUnit)
			fmt.Println(sKey, genericPercepUnit)
			return errors.New("Bad perception unit in Data.KeyToItem: cell is 0 - probably was not matched to a category name or input character")
		}
		data = append(data, percepUnit)
	}
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, []byte(d), os.ModePerm)
}

/*
LoadVocab loads the Charcatnn vocab into Data.KeyToItem so the libpotential training
vocab has input and output values mapped with training data.

potential lib requires generic `map[interface{}]potential.PerceptionUnit`.
*/
func (charcat *Charcatnn) LoadVocab(settings *potential.TrainingSettings, filename string) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	var charcatVocabList []charcatPerceptionUnit
	err = json.Unmarshal(bytes, &charcatVocabList)
	if err != nil {
		return err
	}
	if settings.Data.KeyToItem == nil {
		settings.Data.KeyToItem = make(map[interface{}]potential.PerceptionUnit)
	}
	for index, charcatPercep := range charcatVocabList {
		var genericPercepUnit potential.PerceptionUnit
		if charcatPercep.CategoryName != "" {
			genericPercepUnit = potential.PerceptionUnit{
				Value:      charcatPercep.CategoryName,
				OutputCell: charcatPercep.CellID,
			}
		} else if charcatPercep.InputChar != "" {
			genericPercepUnit = potential.PerceptionUnit{
				Value:     charcatPercep.InputChar,
				InputCell: charcatPercep.CellID,
			}
		} else {
			fmt.Println("array index=", index, charcatPercep)
			return errors.New("Failed loading vocab item for Charcatnn")
		}
		settings.Data.KeyToItem[genericPercepUnit.Value] = genericPercepUnit
	}

	return nil
}

// SetRawData receives the file bytes from the structured training data,
// and parses it into perception units.
func (charcat *Charcatnn) SetRawData(bytes []byte) {
	var charcatVocabList []charcatTrainingData
	err := json.Unmarshal(bytes, &charcatVocabList)
	if err != nil {
		panic(err)
	}
	charcat.structuredTrainingCases = charcatVocabList
}

const maxTries = 50

func addNewVocabMapping(tag string, network *potential.Network) (cellID potential.CellID) {
	for i := 0; i < maxTries; i++ {
		tryCell := network.RandomCellKey()
		isUnused := network.Cells[tryCell].Tag == ""
		if isUnused {
			cellID = tryCell
			break
		}
	}
	if cellID == 0 {
		panic(errors.New("Failed assigning vocab for " + tag + " - likely need a larger network"))
	}

	network.Cells[cellID].Tag = tag
	network.Cells[cellID].Immortal = true

	return cellID
}

/*
PrepareData looks at each character, builds up a map of string:PerceptionUnit pairs.
String will be either the input (single character), or output category (string).

In other words, setup the vocab mappings between 1) input chars and Cells, and 2) output categories and cells.

It also sets up the training samples.

Note: each PerceptionUnit has only an input or an output. This is by design;
it may indicate a flaw in the design of PerceptionUnit.
*/
func (charcat *Charcatnn) PrepareData(settings *potential.TrainingSettings, network *potential.Network) {
	if settings.Data.KeyToItem == nil { // may have been preloaded
		settings.Data.KeyToItem = make(map[interface{}]potential.PerceptionUnit)
	}

	// Make sure all training cases have their components
	// represented by cells in the network

	for _, tc := range charcat.structuredTrainingCases {
		// map the output category to a cell.
		// also make it available to the input chars below, so it
		// can be used to setup training cases.
		var categoryCellID potential.CellID
		categoryPU, categoryExists := settings.Data.KeyToItem[tc.CategoryName]
		if !categoryExists {
			categoryCellID = addNewVocabMapping(tc.CategoryName, network)
			// fmt.Println("adding cat", tc.CategoryName, categoryCellID)
			settings.Data.KeyToItem[tc.CategoryName] = potential.PerceptionUnit{
				Value:      tc.CategoryName,
				OutputCell: categoryCellID,
			}

		} else {
			// fmt.Println("cat exists", categoryPU.Value, categoryPU.OutputCell, categoryPU.InputCell)
			categoryCellID = categoryPU.OutputCell
		}

		// Map each character of the input to a cell for vocab purposes.
		// Then setup training data.
		inputCharsGroup := strings.Split(tc.InputText, "")
		var charCellID potential.CellID
		var samples []*potential.TrainingSample
		for _, inputChar := range inputCharsGroup {
			if _, charVocabExists := settings.Data.KeyToItem[inputChar]; !charVocabExists {
				charCellID = addNewVocabMapping("in-"+inputChar, network)
				settings.Data.KeyToItem[inputChar] = potential.PerceptionUnit{
					Value:     inputChar,
					InputCell: charCellID,
				}
				// fmt.Println("added input", settings.Data.KeyToItem[inputChar])
			} else {
				charCellID = settings.Data.KeyToItem[inputChar].InputCell
			}

			// add training data

			ts := potential.TrainingSample{
				InputCell:  charCellID,
				OutputCell: categoryCellID,
			}

			samples = append(samples, &ts)
		}
		settings.TrainingSamples = append(settings.TrainingSamples, samples)

	}

	// Reverse the mappings we made above
	settings.Data.CellToKey = make(map[potential.CellID]interface{})

	for key, dataItem := range settings.Data.KeyToItem {
		// reverse the map
		if dataItem.InputCell != potential.CellID(0) {
			settings.Data.CellToKey[dataItem.InputCell] = key
		}
		if dataItem.OutputCell != potential.CellID(0) {
			settings.Data.CellToKey[dataItem.OutputCell] = key
		}
	}
}

// SeedAndSample writes the output category to stdout
func (charcat *Charcatnn) SeedAndSample(settings *potential.TrainingSettings, seed []byte, network *potential.Network) {
	seedHexCodeChars := strings.Split(string(seed), "")
	// keys are type interface{} and need to be copied into a new array of that
	// type. they cannot be downcast. https://golang.org/doc/faq#convert_slice_of_interface
	// (might want to add this as a helper)
	var seedKeys []interface{}
	for _, stringKeyChar := range seedHexCodeChars {
		seedKeys = append(seedKeys, stringKeyChar)
	}
	out := potential.Sample(seedKeys, settings.Data, network, 1, nil, nil)
	for _, s := range out {
		fmt.Print(s.(string))
	}
}
