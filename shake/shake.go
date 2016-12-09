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
	"sync"
	"syscall"
	"time"

	"github.com/pkg/profile"
)

const fireCharacterIterations = 4
const initialNetworkNeurons = 200
const defaultNeuronSynapses = 5
const pretrainNeuronsToGrow = 20
const pretrainSynapsesToGrow = 50
const growPathExpectedMinimumSynapses = 10
const linesBetweenPruningSessions = 20
const sleepBetweenInputTriggers = potential.RefractoryPeriodMillis * time.Millisecond
const networkDisabledFizzleOutPeriod = 100 * time.Millisecond

var networkSaveFile = flag.String("save", "network.json", "Load/save location of the network")
var vocabSaveFile = flag.String("vocab", "vocab.json", "Load/save location of the charrnn vocab")
var testDataFile = flag.String("data", "shake.txt", "File location of the training data.")
var train = flag.Int("train", 0, "Train the network with this number of workers")
var seed = flag.String("seed", "", "Seed the neural network with this text then sample it.")
var doProfile = flag.String("profile", "", "Pass `cpu` or `mem` to do profiling")

func main() {
	// doTrace()

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
		fmt.Println("Created network")
		fmt.Println("Saving to disk")
	} else {
		fmt.Println("Loaded network from disk,", len(network.Cells), "cells",
			len(network.Synapses), "synapses")
	}

	bytes, err := ioutil.ReadFile(*testDataFile)
	if err != nil {
		panic(err)
	}

	text := string(bytes)
	lines := strings.Split(text, "\n")
    settings := potential.NewTrainingSettings()

	t := charrnn.Charrnn{
        Chars: lines,
        Settings: settings
    }
    err := t.LoadVocab(*vocabSaveFile)
    if err != nil {
        t.PrepareData(network)
    }

	fmt.Println("Loaded training text for", *testDataFile, "length=", len(c.Settings.Data.KeyToItem))

	// Figure out how to run this program.
	flag.Parse()
	if len(*seed) > 0 {
		fmt.Println("Sampling characters with seed text: ", *seed)
		potential.Sample(*seed, data, network, 1000, "START", "END")
		return
	}
	t.Settings.Threads := *train
	if t.Settings.Threads == 0 {
		fmt.Println("Not enough params. Help:")
		flag.PrintDefaults()
		fmt.Println("")
		return
	}
	// only profile during training
	if *doProfile == "mem" {
		defer profile.Start(profile.MemProfile).Stop()
	} else if *doProfile == "cpu" {
		defer profile.Start(profile.CPUProfile).Stop()
	}

	network.Disabled = true // we just will never need it to fire

	fmt.Println("Beginning training", threads, "simultaneous sessions")

	// make sure we save any progress from a long running training
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		now := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
		err = network.SaveToFile("network_" + now + ".json")
		if err != nil {
			fmt.Println(err)
		}
		os.Exit(1)
	}()

	err = potential.Train(t, t.Settings, network)
	if err != nil {
		fmt.Println("Training failed:", err)
	}

	err = network.SaveToFile(*networkSaveFile)
	if err != nil {
		fmt.Println("Failed saving network")
		fmt.Println(err)
		return
	}
	fmt.Println("Done.")
}
