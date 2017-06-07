package potential

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
)

/*
Vocabulary holds the input and output values as well as some training samples.
*/
type Vocabulary struct {
	Net *Network `json:"-"`
	/*
		Inputs are low level units of input that map to a bunch of input
		cells. Sort of like how the letter A might fire a group of neurons
		in a human brain.
	*/
	Inputs     map[InputValue]*VocabUnit
	Outputs    map[OutputValue]*OutputCollection
	Samples    []sample
	Threads    int
	Workerfile string
}

/*
NewVocabulary is a factory
*/
func NewVocabulary(network *Network) *Vocabulary {
	return &Vocabulary{
		Net:     network,
		Inputs:  make(map[InputValue]*VocabUnit),
		Outputs: make(map[OutputValue]*OutputCollection),
		Threads: runtime.NumCPU(),
	}
}

// sample is a training data sample
type sample struct {
	input  InputValue
	output OutputValue
}

/*
ClearSamples clears the samples
*/
func (vocab *Vocabulary) ClearSamples() {
	vocab.Samples = make([]sample, 0)
}

/*
CloneOutputs returns a new grouping of NEW OutputCollections (not shared
pointers).
*/
func CloneOutputs(outputs map[OutputValue]*OutputCollection) map[OutputValue]*OutputCollection {
	newOutputs := make(map[OutputValue]*OutputCollection)

	for k, v := range outputs {
		newOutputs[k] = &OutputCollection{
			Value:       v.Value,
			FirePattern: v.FirePattern,
		}
	}
	return outputs
}

/*
AddTrainingData takes a group of units, such as a group of
pixels, or a word, and breaks it into its smaller parts. Then it finds those
corresponding smaller parts in the VocabUnit collection. It adds training
samples for this also
*/
func (vocab *Vocabulary) AddTrainingData(unitGroups []interface{}) {
	var lastChar string
	for _, inputGroup := range unitGroups {
		groupParts := strings.Split(fmt.Sprint(inputGroup), "")

		// make sure there is an input for this character
		for _, char := range groupParts {
			_, exists := vocab.Inputs[InputValue(char)]
			if !exists {
				vu := NewVocabUnit(char)
				vu.InitRandomInputs(vocab.Net)
				vocab.Inputs[InputValue(char)] = vu
			}
			firstHasNoPreceedingPredictor := lastChar == ""
			if firstHasNoPreceedingPredictor {
				lastChar = char
				continue
			}
			// preceeding group predicts this one
			vocab.Samples = append(vocab.Samples, sample{
				input:  InputValue(lastChar),
				output: OutputValue(char),
			})
		}
	}
}

/*
LoadVocabFromFile loads the vocab from a JSON file but does not populate
the Net (the network).
*/
func LoadVocabFromFile(filepath string) (vocab *Vocabulary, err error) {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return vocab, err
	}
	err = json.Unmarshal(bytes, &vocab)
	return vocab, err
}

/*
SaveToFile saves the vocab to a JSON file but does not save the Net
property (the network).
*/
func (vocab *Vocabulary) SaveToFile(filepath string) error {
	d, err := json.Marshal(vocab)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath, d, os.ModePerm)
}

func (vocab *Vocabulary) printInputs() {
	for k, v := range vocab.Inputs {
		fmt.Println("Input:", k, v)
	}
}

func (vocab *Vocabulary) printOutputs() {
	for k, v := range vocab.Outputs {
		fmt.Println("Output:", k, v)
	}
}
