package potential

import (
	"encoding/json"
	"fmt"
	"strings"
)

/*
Sample produces the raw string output based on seed text that was input
by the user.
*/
func Sample(seedText string, vocab *Vocabulary) (output string) {
	characters := strings.Split(string(seedText), "")
	var charArray []interface{}
	for _, char := range characters {
		charArray = append(charArray, interface{}(char))
	}
	// do some shenanigans to get data in the right format
	unit := UnitGroup{InputText: seedText}
	unitArray := make([]*UnitGroup, 1)
	unitArray[0] = &unit
	unitJSON, err := json.Marshal(unitArray)
	if err != nil {
		fmt.Println(
			"Failed parsing UnitGroup with seedText into JSON. seedText=",
			seedText)
		panic(err)
	}

	err = vocab.AddTrainingData(unitJSON)
	if err != nil {
		panic(err)
	}
	output = ""
	// fire the samples, not resetting in between (?)
	for _, s := range vocab.Samples {
		// fire the input a bunch of times. after that we can consider
		// the output pattern as fired. set the output pattern.
		cellsToFireForInputValues := GetInputPatternForInputs(vocab, s.inputs)
		finalPattern := FireNetworkUntilDone(vocab.Net, cellsToFireForInputValues)
		closest := FindClosestOutputCollection(finalPattern, vocab)
		if closest != nil {
			output += string(closest.Value)
		}
	}

	// collect the firing pattern at each step

	// find closest firing pattern and append its output to the sample

	// once all inputs fired and network fizzles out, end the sampling

	return output
}
