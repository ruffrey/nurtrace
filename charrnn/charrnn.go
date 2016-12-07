package charrnn

import (
	"bleh/potential"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

/*
Charrnn is the collection of training stuff to operate upon

It implements potential.Trainer
*/
type Charrnn struct {
	chars    []string
	Settings *potential.TrainingSettings
}

/*
SaveVocab saves the current vocabulary from the charrnn.
*/
func (charrnn *Charrnn) SaveVocab(filename string) error {
	data, err := json.Marshal(charrnn.Settings.Data.KeyToItem)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, []byte(data), os.ModePerm)
}

/*
LoadVocab loads the known vocabulary and mappings to cells and puts it in
the settings.
*/
func (charrnn *Charrnn) LoadVocab(filename string) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, charrnn.Settings.Data)
	return err
}

/*
PrepareData is from potential.Trainer
*/
func (charrnn Charrnn) PrepareData(network *potential.Network) {
	fmt.Println(charrnn.Settings)
	charrnn.Settings.Data.KeyToItem = make(map[interface{}]potential.PerceptionUnit)

	// from the training samples, build the vocabulary
	for _, Value := range charrnn.chars {
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
}

/*
OnTrained is the callback from potential.Trainer
*/
func (charrnn Charrnn) OnTrained() {
	fmt.Println("charrnn done")
}
