package potential

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	Noise      FiringPattern
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
		Noise:   make(FiringPattern),
	}
}

// sample is a training data sample
type sample struct {
	inputs []InputValue
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
UnitGroup is a basic training data item. A JSON array of these makes
TrainingData.
*/
type UnitGroup struct {
	InputText      string
	ExpectedOutput string
}

/*
TrainingData consists of an array of UnitGroups, and these would be
stored as normal JSON. Users must supply this training data.
*/
type TrainingData []*UnitGroup

/*
AddTrainingData takes a group of units, such as a group of
pixels, or a word, and breaks it into its smaller parts. Then it finds those
corresponding smaller parts in the VocabUnit collection. It adds training
samples for this also
*/
func (vocab *Vocabulary) AddTrainingData(testDataBytes []byte) (err error) {
	// for now, convert to string
	td := make(TrainingData, 0)
	err = json.Unmarshal(testDataBytes, &td)
	if err != nil {
		log.Println("Unable to parse training data JSON", err)
		return err
	}

	for _, inputGroup := range td {
		inputParts := strings.Split(fmt.Sprint(inputGroup.InputText), "")
		var inputs []InputValue
		output := OutputValue(inputGroup.ExpectedOutput)

		// make sure there is an input for this character
		for _, char := range inputParts {
			_, exists := vocab.Inputs[InputValue(char)]
			if !exists {
				vu := NewVocabUnit(char)
				vu.InitRandomInputs(vocab)
				vocab.Inputs[InputValue(char)] = vu
			}
			inputs = append(inputs, InputValue(char))
		}

		// make sure output exists
		if _, exists := vocab.Outputs[output]; !exists {
			vocab.Outputs[output] = NewOutputCollection(output)
		}

		vocab.Samples = append(vocab.Samples, sample{
			inputs,
			output,
		})

	}

	return nil
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
		log.Println("Input:", k, v)
	}
}

func (vocab *Vocabulary) printOutputs() {
	for k, v := range vocab.Outputs {
		log.Println("Output:", k, v)
	}
}
