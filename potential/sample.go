package potential

import (
	"strings"
)

/*
Sample produces the raw string output based on seed text that was input
by the user.
*/
func Sample(seedText string, vocab *Vocabulary, maxLength int) (output string) {
	characters := strings.Split(string(seedText), "")
	charTotal := len(characters)
	inputs := make([]InputValue, charTotal)
	for i := 0; i < charTotal; i++ {
		inputs[i] = InputValue(characters[i])
	}

	output = ""
	vocab.Net.ResetForTraining()
	// need to combine cells to be fired
	cellsToFireForInputValues := make(FiringPattern)

	fp := GetInputPatternForInputs(vocab, inputs)
	cellsToFireForInputValues = mergeFiringPatterns(cellsToFireForInputValues, fp)
	finalPattern := FireNetworkUntilDone(vocab.Net, cellsToFireForInputValues)

	// TODO: find more than one match?
	closest := FindClosestOutputCollection(finalPattern, vocab)
	if closest != nil {
		output += string(closest.Value)
	}

	// TODO: this is useless code
	if len(output) >= maxLength {
		return output
	}

	return output
}
