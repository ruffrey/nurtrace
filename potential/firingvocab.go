package potential

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

/*
Vocabulary holds the input and output values as well as some training samples.
*/
type Vocabulary struct {
	Net     *Network `json:"-"`
	Inputs  map[InputValue]*VocabUnit
	Outputs map[OutputValue]*OutputCollection
	samples map[InputValue]OutputValue
}

/*
addTrainingData takes a group of units, such as a group of
pixels, or a word, and breaks it into its smaller parts. Then it finds those
corresponding smaller parts in the VocabUnit collection. It adds training
samples for
*/
func (vocab *Vocabulary) addTrainingData(unitGroups []interface{}, expected string) {
	for ix, inputGroup := range unitGroups {
		groupParts := strings.Split(fmt.Sprint(inputGroup), "")

		// make sure there is an input for this character
		for _, char := range groupParts {
			_, exists := vocab.Inputs[InputValue(char)]
			if !exists {
				vu := NewVocabUnit(char)
				vu.InitRandomInputs(vocab.Net)
				vocab.Inputs[InputValue(char)] = vu
			}
		}

		firstHasNoPreceedingPredictor := ix == 0
		if firstHasNoPreceedingPredictor {
			continue
		}
		// preceeding group predicts this one
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
	return ioutil.WriteFile(filepath, []byte(d), os.ModePerm)
}
