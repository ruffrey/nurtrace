package potential

import (
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
	// produce a set of samples - []sample
	vocab.AddTrainingData(charArray)

	// fire the samples, not resetting in between (?)
	for _, s := range vocab.Samples {
		// fire the input a bunch of times. after that we can consider
		// the output pattern as fired. set the output pattern.
		inputs := vocab.Inputs[s.input].InputCells
		finalPattern := FireNetworkUntilDone(vocab.Net, inputs)
		closest := FindClosestOutputCollection(finalPattern, vocab)
		output += string(closest.Value)
	}

	// collect the firing pattern at each step

	// find closest firing pattern and append its output to the sample

	// once all inputs fired and network fizzles out, end the sampling

	return output
}
