package potential

import (
	"fmt"
	"strings"
)

/*
Vocabulary holds the input and output values as well as some training samples.
*/
type Vocabulary struct {
	Net     *Network
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
