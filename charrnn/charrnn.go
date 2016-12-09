package charrnn

// TODO: saving/restoring from disk does not work.

import (
	"bleh/potential"
	"encoding/json"
	"io/ioutil"
	"os"
)

/*
Charrnn is the collection of training stuff to operate upon

It implements potential.Trainer
*/
type Charrnn struct {
	Chars    []string
	Settings *potential.TrainingSettings
}

/*
charPerceptionUnit is almost the same as a PerceptionUnit but the value is typed to string,
which enables it to become json or get serialized back into json.
*/
type charPerceptionUnit struct {
	Value      string
	InputCell  potential.CellID
	OutputCell potential.CellID
}

/*
SaveVocab saves the current vocabulary from the charrnn.
*/
func (charrnn *Charrnn) SaveVocab(filename string) error {
	data := make(map[string]charPerceptionUnit)
	for key, value := range charrnn.Settings.Data.KeyToItem {
		data[key.(string)] = charPerceptionUnit{
			Value:      value.Value.(string),
			InputCell:  value.InputCell,
			OutputCell: value.OutputCell,
		}
	}
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, []byte(d), os.ModePerm)
}

/*
LoadVocab loads the known vocabulary and mappings to cells and puts it in
the charrnn.Settings. It uses the charPerceptionUnit as an intermediary but
casts it back into a generic `map[interface{}]potential.PerceptionUnit`, which
the potential lib requires.
*/
func (charrnn *Charrnn) LoadVocab(filename string) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	data := make(map[string]charPerceptionUnit)
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return err
	}
	if charrnn.Settings.Data.KeyToItem == nil {
		charrnn.Settings.Data.KeyToItem = make(map[interface{}]potential.PerceptionUnit)
	}
	for key, value := range data {
		charrnn.Settings.Data.KeyToItem[key] = potential.PerceptionUnit{
			Value:      value.Value,
			InputCell:  value.InputCell,
			OutputCell: value.OutputCell,
		}
	}

	return nil
}

/*
PrepareData is from potential.Trainer. Looking at each character, build up
a map of string: PerceptionUnit pairs.
*/
func (charrnn Charrnn) PrepareData(network *potential.Network) {
	if charrnn.Settings.Data.KeyToItem == nil { // may have been preloaded
		charrnn.Settings.Data.KeyToItem = make(map[interface{}]potential.PerceptionUnit)
	}

	// From the characters, ensure the vocabulary is all setup.
	// Nothing will change in the model if this all characters already have corresponding
	// cells, i.e. this is not the first time training the network.
	for _, Value := range charrnn.Chars {
		if _, exists := charrnn.Settings.Data.KeyToItem[Value]; !exists {
			InputCell := network.RandomCellKey()
			// ensure the input and output cells are not the same!
			var OutputCell potential.CellID
			for {
				OutputCell = network.RandomCellKey()
				if OutputCell != InputCell {
					break
				}
			}

			charrnn.Settings.Data.KeyToItem[Value] = potential.PerceptionUnit{
				Value:      Value,
				InputCell:  InputCell,
				OutputCell: OutputCell,
			}
		}
	}

	// Again, if this is the first time training the network, we must setup start
	// and end indicators.
	start := potential.PerceptionUnit{
		Value:     "START",
		InputCell: network.RandomCellKey(),
	}
	for {
		start.OutputCell = network.RandomCellKey()
		if start.InputCell != start.OutputCell {
			break
		}
	}
	end := potential.PerceptionUnit{
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

	charrnn.Settings.Data.KeyToItem["START"] = start
	charrnn.Settings.Data.KeyToItem["END"] = end

	// add data
	charrnn.Settings.Data.CellToKey = make(map[potential.CellID]interface{})

	for key, dataItem := range charrnn.Settings.Data.KeyToItem {
		// grow paths between all the inputs and outputs
		network.GrowPathBetween(dataItem.InputCell, dataItem.OutputCell, 10)
		// reverse the map
		charrnn.Settings.Data.CellToKey[dataItem.InputCell] = key
		charrnn.Settings.Data.CellToKey[dataItem.OutputCell] = key
		// prevent accidentally pruning the input/output cells
		network.Cells[dataItem.InputCell].Immortal = true
		network.Cells[dataItem.OutputCell].Immortal = true
	}
}
