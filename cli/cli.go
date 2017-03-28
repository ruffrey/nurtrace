package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ruffrey/nurtrace/laws"
	"github.com/ruffrey/nurtrace/perception"
	"github.com/ruffrey/nurtrace/perceptions/charrnn"
	"github.com/ruffrey/nurtrace/potential"

	"github.com/pkg/profile"
)

const percepModList = "charrnn, category"

var perceptionModel = flag.String("model", "", "Perception model type: "+percepModList)
var networkSaveFile = flag.String("save", "network.json", "Load/save location of the network file")
var vocabSaveFile = flag.String("vocab", "vocab.json", "Load/save location of the vocab mapping file")
var testDataFile = flag.String("data", "", "File location of the training data.")
var seed = flag.String("seed", "", "Seed the neural network with this data then sample it.")
var doProfile = flag.String("profile", "", "Pass `cpu` or `mem` to do profiling")
var initialNetworkNeurons = flag.Int("startsize", 200, "Start size of network when creating a new one")

func main() {
	// Figure out how they want to run this program.
	flag.Parse()

	var t perception.Perception
	switch *perceptionModel {
	case "category":
		// t = category.Category.New()
	case "charrnn":
		m := charrnn.Charrnn{}
		t = &m
		break
	default:
		fmt.Println("Perception -model flag valid choices are:", percepModList)
		flag.PrintDefaults()
		return
	}
	settings := potential.NewTrainingSettings()
	// start by initializing the network from disk or whatever
	var network *potential.Network
	var err error
	network, err = potential.LoadNetworkFromFile(*networkSaveFile)
	if err != nil {
		fmt.Println("Unable to load network from file; creating new one.", err)
		network = potential.NewNetwork()
		neuronsToAdd := *initialNetworkNeurons
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

	// t.Settings.Workerfile = "Workerfile"
	t.SetRawData(bytes)
	err = t.LoadVocab(settings, *vocabSaveFile)
	t.PrepareData(settings, network)

	fmt.Println("Loaded training text for", *testDataFile, "samples=", len(settings.TrainingSamples))

	// Sample, then stop.
	if len(*seed) > 0 {
		t.PrepareData(settings, network) // make sure all data is setup
		t.SeedAndSample(settings, *seed, network)
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
		err = t.SaveVocab(settings, *vocabSaveFile)
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
	err = t.SaveVocab(settings, *vocabSaveFile)
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
