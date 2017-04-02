package cmd

import (
	"errors"

	"github.com/ruffrey/nurtrace/perception"
	"github.com/ruffrey/nurtrace/perceptions/charcat"
	"github.com/ruffrey/nurtrace/perceptions/charrnn"
	"github.com/ruffrey/nurtrace/potential"
)

// Sample uses a pretrained network to generate a prediction based on user provided data.
func Sample(perceptionModel, networkSaveFile, vocabSaveFile string, seed []byte) (err error) {
	// Figure out how they want to run this program.

	var t perception.Perception
	switch perceptionModel {
	case "category":
		m := charcat.Charcatnn{}
		t = &m
		break
	case "charrnn":
		m := charrnn.Charrnn{}
		t = &m
		break
	default:
		return errors.New("Perception model valid choices are: charrnn, category")
	}
	settings := potential.NewTrainingSettings()
	// start by initializing the network from disk or whatever
	var network *potential.Network
	network, err = potential.LoadNetworkFromFile(networkSaveFile)
	if err != nil {
		return err
	}

	err = t.LoadVocab(settings, vocabSaveFile)
	if err != nil {
		return err
	}

	t.PrepareData(settings, network) // make sure all data is setup
	t.SeedAndSample(settings, seed, network)

	return nil
}
