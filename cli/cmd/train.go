package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/pkg/profile"
	"github.com/ruffrey/nurtrace/laws"
	"github.com/ruffrey/nurtrace/perception"
	"github.com/ruffrey/nurtrace/perceptions/charcat"
	"github.com/ruffrey/nurtrace/perceptions/charrnn"
	"github.com/ruffrey/nurtrace/potential"
)

// Train performs training of a new or existing network and saves it to disk.
func Train(perceptionModel, networkInputFile, networkSaveFile, vocabSaveFile, testDataFile, doProfile string, initialNetworkNeurons int) (err error) {
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
		fmt.Println(err)
		fmt.Println("Unable to load network from file; creating new one.")
		network = potential.NewNetwork()
		neuronsToAdd := initialNetworkNeurons
		synapsesToAdd := 0
		network.Grow(neuronsToAdd, laws.DefaultNeuronSynapses, synapsesToAdd)
		fmt.Println("Created network,", len(network.Cells), "cells",
			len(network.Synapses), "synapses")
	} else {
		fmt.Println("Loaded network from disk")
		network.PrintTotals()
	}

	fmt.Println("Reading test data file", testDataFile)
	bytes, err := ioutil.ReadFile(testDataFile)
	if err != nil {
		fmt.Println("Unable to read test data file", testDataFile)
		return err
	}

	// t.Settings.Workerfile = "Workerfile"
	fmt.Println("Setting raw data")
	t.SetRawData(bytes)
	fmt.Println("Attempting to load or create vocab")
	err = t.LoadVocab(settings, vocabSaveFile)
	fmt.Println("Preparing data")
	t.PrepareData(settings, network)

	fmt.Println("Loaded training data for", testDataFile, "- samples=", len(settings.TrainingSamples))

	// only profile during training
	if doProfile == "mem" {
		defer profile.Start(profile.MemProfile).Stop()
	} else if doProfile == "cpu" {
		defer profile.Start(profile.CPUProfile).Stop()
	}

	// Make sure we save any progress from a long running training that the user ends.
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		now := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
		err = t.SaveVocab(settings, vocabSaveFile)
		if err != nil {
			fmt.Println("Failed saving vocab")
			fmt.Println(err)
		}
		err = network.SaveToFile("network_" + now + ".json")
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(1)
	}()

	fmt.Println("Beginning training")
	network.Disabled = true // we just will never need it to fire
	potential.Train(settings, network, "")

	// Training is over

	// Ensure we save the vocab
	err = t.SaveVocab(settings, vocabSaveFile)
	if err != nil {
		fmt.Println("Failed saving vocab")
		fmt.Println(err)
	}
	// Save the network
	err = network.SaveToFile(networkSaveFile)
	if err != nil {
		fmt.Println("Failed saving network")
		fmt.Println(err)
	}

	network.PrintTotals()
	fmt.Println("Done.")

	return nil
}
