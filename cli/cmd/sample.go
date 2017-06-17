package cmd

import (
	"log"

	"github.com/ruffrey/nurtrace/potential"
)

// Sample uses a pretrained network to generate a prediction based on user provided data.
func Sample(networkSaveFile, vocabSaveFile string, seedText string, sampleLength int) (err error) {
	var vocab *potential.Vocabulary
	vocab, err = potential.LoadVocabFromFile(vocabSaveFile)
	if err != nil {
		return err
	}
	var network *potential.Network
	network, err = potential.LoadNetworkFromFile(networkSaveFile)
	if err != nil {
		return err
	}

	vocab.Net = network
	output := potential.Sample(seedText, vocab, sampleLength)

	log.Println(output)

	return nil
}
