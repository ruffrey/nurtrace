package charrnn

// TODO: saving/restoring from disk does not work.

import (
	"bleh/perception"
	"bleh/potential"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

/*
Charrnn is the collection of training stuff to operate upon

It implements potential.Trainer
*/
type Charrnn struct {
	Chars    []string
	rawText  string
	settings *potential.TrainingSettings
	perception.Perception
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
func (charrnn Charrnn) SaveVocab(filename string) error {
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
func (charrnn Charrnn) LoadVocab(filename string) error {
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

// SetRawData sets the rawText prop
func (charrnn Charrnn) SetRawData(bytes []byte) {
	text := string(bytes)
	charrnn.rawText = text
	charrnn.Chars = strings.Split(text, "")
}

// SeedAndSample writes the output sample to stdout at the moment
func (charrnn Charrnn) SeedAndSample(seed string, network *potential.Network) {
	fmt.Println("Sampling characters with seed text: ", seed)
	seedChars := strings.Split(seed, "")
	// keys are type interface{} and need to be copied into a new array of that
	// type. they cannot be downcast. https://golang.org/doc/faq#convert_slice_of_interface
	// (might want to add this as a helper to charrnn)
	var seedKeys []interface{}
	for _, stringKeyChar := range seedChars {
		seedKeys = append(seedKeys, stringKeyChar)
	}
	out := potential.Sample(seedKeys, charrnn.Settings.Data, network, 1000, "START", "END")
	fmt.Println("---")
	for _, s := range out {
		fmt.Print(s)
	}
	fmt.Println("\n---")
}

/*
PrepareData looks at each character, builds up a map of string: PerceptionUnit pairs.
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
			network.Cells[InputCell].Tag = "in-" + Value
			network.Cells[OutputCell].Tag = "out-" + Value
		}
	}

	// Again, if this is the first time training the network, we must setup start
	// and end indicators.

	if _, exists := charrnn.Settings.Data.KeyToItem["START"]; !exists {
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

	// Reverse the map and grow paths where needed
	charrnn.Settings.Data.CellToKey = make(map[potential.CellID]interface{})

	for key, dataItem := range charrnn.Settings.Data.KeyToItem {
		// reverse the map
		charrnn.Settings.Data.CellToKey[dataItem.InputCell] = key
		charrnn.Settings.Data.CellToKey[dataItem.OutputCell] = key
		// prevent accidentally pruning the input/output cells
		network.Cells[dataItem.InputCell].Immortal = true
		network.Cells[dataItem.OutputCell].Immortal = true
		// Right here we used to grow a path between the input and
		// the output. But that is utterly incorrect - it created a
		// path between, say, A and A, b and b, c and c, etc.
	}

	// Setup the training data samples
	//
	// One batch will be one line, with pairs being start-<line text>-end
	lines := strings.Split(charrnn.rawText, "\n")
	startCellID := charrnn.Settings.Data.KeyToItem["START"].InputCell
	endCellID := charrnn.Settings.Data.KeyToItem["END"].InputCell
	for _, line := range lines {
		var s []*potential.TrainingSample
		chars := strings.Split(line, "")

		if len(chars) == 0 {
			continue
		}
		// first char is START indicator token
		ts1 := potential.TrainingSample{
			InputCell:  startCellID,
			OutputCell: charrnn.Settings.Data.KeyToItem[chars[0]].InputCell,
		}
		if ts1.InputCell == 0 {
			fmt.Println(ts1)
			panic("nope")
		}
		s = append(s, &ts1)

		// start at 1 because we need to look behind
		for i := 1; i < len(chars); i++ {
			ts := potential.TrainingSample{
				InputCell:  charrnn.Settings.Data.KeyToItem[chars[i-1]].InputCell,
				OutputCell: charrnn.Settings.Data.KeyToItem[chars[i]].InputCell,
			}
			if ts.InputCell == 0 {
				fmt.Println("training sample has input cell where ID input cell is zero")
				fmt.Println(i, chars[i], chars[i-1], ts)
				panic(i)
			}
			s = append(s, &ts)
		}
		// last char is END indicator token
		ts2 := potential.TrainingSample{
			InputCell:  charrnn.Settings.Data.KeyToItem[chars[len(chars)-1]].InputCell,
			OutputCell: endCellID,
		}
		if ts2.InputCell == 0 {
			fmt.Println(ts2)
			fmt.Println("training sample END has input cell where ID input cell is zero")
			panic("nope")
		}
		s = append(s, &ts2)

		charrnn.Settings.TrainingSamples = append(charrnn.Settings.TrainingSamples, s)
	}

	fmt.Println("charrnn data setup complete")
}
