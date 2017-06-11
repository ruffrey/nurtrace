package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/pkg/profile"
	"github.com/ruffrey/nurtrace/laws"
	"github.com/ruffrey/nurtrace/potential"
)

// Train trains a network and vocab set.
func Train(networkInputFile, networkSaveFile, vocabSaveFile, testDataFile, doProfile string, initialNetworkNeurons int) (err error) {
	// start by initializing the network from disk or whatever
	var network *potential.Network
	var vocab *potential.Vocabulary

	// Load network
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

	// Load vocab
	vocab, err = potential.LoadVocabFromFile(vocabSaveFile)
	if err != nil {
		fmt.Println(err)
		vocab = potential.NewVocabulary(network)
		fmt.Println("Created vocab", vocabSaveFile)
	} else {
		fmt.Println("Loaded vocab from disk", vocabSaveFile)
	}
	vocab.Net = network

	fmt.Println("Reading training data file", testDataFile)
	testDataBytes, err := ioutil.ReadFile(testDataFile)
	if err != nil {
		fmt.Println("Unable to read training data file", testDataFile, err)
		return err
	}
	err = vocab.AddTrainingData(testDataBytes)
	if err != nil {
		fmt.Println("Failed adding training data", testDataFile, err)
		return err
	}

	// TODO: Workerfile

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
		err = vocab.SaveToFile("vocab_" + now + ".json")
		if err != nil {
			fmt.Println("Failed saving vocab")
			fmt.Println(err)
		}
		err = network.SaveToFile("network_" + now + ".nur")
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(1)
	}()

	fmt.Println("Beginning training")
	network.Disabled = true // we just will never need it to fire
	potential.Train(vocab, "")

	// Training is over

	// Ensure we save the vocab, but empty the samples first.
	vocab.ClearSamples()
	err = vocab.SaveToFile(vocabSaveFile)
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
