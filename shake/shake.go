package main

import (
	"bleh/laws"
	"bleh/perception"
	"bleh/potential"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/pkg/profile"
)

const initialNetworkNeurons = 200

var networkSaveFile = flag.String("save", "network.json", "Load/save location of the network")
var vocabSaveFile = flag.String("vocab", "vocab.json", "Load/save location of the charrnn vocab")
var testDataFile = flag.String("data", "shake.txt", "File location of the training data.")
var seed = flag.String("seed", "", "Seed the neural network with this text then sample it.")
var doProfile = flag.String("profile", "", "Pass `cpu` or `mem` to do profiling")

func main() {
	// doTrace()

	// Figure out how they want to run this program.
	flag.Parse()

	// start by initializing the network from disk or whatever
	var network *potential.Network
	var err error
	network, err = potential.LoadNetworkFromFile(*networkSaveFile)
	if err != nil {
		fmt.Println("Unable to load network from file; creating new one.", err)
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

	fmt.Println("Reading test data file", *testDataFile)
	bytes, err := ioutil.ReadFile(*testDataFile)
	if err != nil {
		fmt.Println("Unable to read test data file", *testDataFile)
		panic(err)
	}

	settings := potential.NewTrainingSettings()
	t := perception.Perception{
		Settings: settings,
	}
	// t.Settings.Workerfile = "Workerfile"
	t.SetRawData(bytes)
	err = t.LoadVocab(*vocabSaveFile)
	t.PrepareData(network)

	fmt.Println("Loaded training text for", *testDataFile, "samples=", len(t.Settings.TrainingSamples))

	// Sample, then stop.
	if len(*seed) > 0 {
		t.PrepareData(network) // make sure all data is setup
		t.SeedAndSample(*seed, network)
		return
	}

	// only profile during training
	if *doProfile == "mem" {
		defer profile.Start(profile.MemProfile).Stop()
	} else if *doProfile == "cpu" {
		defer profile.Start(profile.CPUProfile).Stop()
	}

	// Make sure we save any progress from a long running training that the user ends.
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		now := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
		err = t.SaveVocab(*vocabSaveFile)
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
	potential.Train(t.GetSettings(), network, "")

	// Training is over

	// Ensure we save the vocab
	err = t.SaveVocab(*vocabSaveFile)
	if err != nil {
		fmt.Println("Failed saving vocab")
		fmt.Println(err)
	}
	// Save the network
	err = network.SaveToFile(*networkSaveFile)
	if err != nil {
		fmt.Println("Failed saving network")
		fmt.Println(err)
		return
	}

	network.PrintTotals()
	fmt.Println("Done.")
}
