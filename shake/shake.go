package main

import (
	"bleh/charrnn"
	"bleh/potential"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/profile"
)

const initialNetworkNeurons = 1000
const defaultNeuronSynapses = 5

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
		newN := potential.NewNetwork()
		network = &newN
		neuronsToAdd := initialNetworkNeurons
		synapsesToAdd := 0
		network.Grow(neuronsToAdd, defaultNeuronSynapses, synapsesToAdd)
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

	text := string(bytes)
	lines := strings.Split(text, "\n")
	chars := strings.Split(text, "")
	settings := potential.NewTrainingSettings()
	// TODO: lines need to be setup for batches of training data.
	t := charrnn.Charrnn{
		Chars:    chars,
		Settings: settings,
	}
	err = t.LoadVocab(*vocabSaveFile)
	t.PrepareData(network)

	// Setup the training data samples
	//
	// One batch will be one line, with pairs being start-<line text>-end
	startCellID := settings.Data.KeyToItem["START"].InputCell // is that right? no?
	endCellID := settings.Data.KeyToItem["END"].InputCell     // is that right? no?
	for _, line := range lines {
		var s []*potential.TrainingSample
		chars := strings.Split(line, "")

		if len(chars) == 0 {
			continue
		}
		// first char is START indicator token
		ts1 := potential.TrainingSample{
			InputCell:  startCellID,
			OutputCell: settings.Data.KeyToItem[chars[0]].InputCell,
		}
		if ts1.InputCell == 0 {
			fmt.Println(ts1)
			panic("nope")
		}
		s = append(s, &ts1)

		// start at 1 because we need to look behind
		for i := 1; i < len(chars); i++ {
			ts := potential.TrainingSample{
				InputCell:  settings.Data.KeyToItem[chars[i-1]].InputCell,
				OutputCell: settings.Data.KeyToItem[chars[i]].InputCell,
			}
			if ts.InputCell == 0 {
				fmt.Println("training sample has input cell where ID input cell is zero")
				fmt.Println(i, chars[i], chars[i-1], ts)
				panic(i)
			}
			s = append(s, &ts)
		}
		// last char is END indicator token
		ts2 := potential.TrainingSample{
			InputCell:  settings.Data.KeyToItem[chars[len(chars)-1]].InputCell,
			OutputCell: endCellID,
		}
		if ts2.InputCell == 0 {
			fmt.Println(ts2)
			fmt.Println("training sample END has input cell where ID input cell is zero")
			panic("nope")
		}
		s = append(s, &ts2)

		settings.TrainingSamples = append(settings.TrainingSamples, s)
	}

	fmt.Println("Loaded training text for", *testDataFile, "samples=", len(t.Settings.TrainingSamples))

	// Sample, then stop.
	if len(*seed) > 0 {
		t.PrepareData(network) // make sure all data is setup
		fmt.Println("Sampling characters with seed text: ", *seed)
		seedChars := strings.Split(*seed, "")
		// keys are type interface{} and need to be copied into a new array of that
		// type. they cannot be downcast. https://golang.org/doc/faq#convert_slice_of_interface
		// (might want to add this as a helper to charrnn)
		var seedKeys []interface{}
		for _, stringKeyChar := range seedChars {
			seedKeys = append(seedKeys, stringKeyChar)
		}
		out := potential.Sample(seedKeys, t.Settings.Data, network, 1000, "START", "END")
		fmt.Println("---")
		for _, s := range out {
			fmt.Print(s)
		}
		fmt.Println("\n---")
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
	potential.Train(t, t.Settings, network)

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
