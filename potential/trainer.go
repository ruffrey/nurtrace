package potential

import "time"

const defaultWorkerThreads = 2
const initialNetworkNeurons = 200
const defaultNeuronSynapses = 5
const pretrainNeuronsToGrow = 20
const pretrainSynapsesToGrow = 50
const growPathExpectedMinimumSynapses = 10
const linesBetweenPruningSessions = 20
const sleepBetweenInputTriggers = RefractoryPeriodMillis * time.Millisecond
const networkDisabledFizzleOutPeriod = 100 * time.Millisecond

/*
TrainingSettings are
*/
type TrainingSettings struct {
	Threads int
	Data    *Dataset
	// List of arrays of cells to fire for training.
	TrainingSamples []map[CellID]CellID
}

/*
NewTrainingSettings sets up and returns a new TrainingSettings instance.
*/
func NewTrainingSettings() *TrainingSettings {
	d := Dataset{}
	dataset := &d
	settings := TrainingSettings{
		Threads:         defaultWorkerThreads,
		Data:            dataset,
		TrainingSamples: make([]map[CellID]CellID, 0),
	}
	return &settings
}

/*
PerceptionUnit is the smallest and core unit of a Dataset.
*/
type PerceptionUnit struct {
	Value      interface{}
	InputCell  CellID
	OutputCell CellID
}

/*
Dataset helps represent the smallest units of a trainable set of data, and maps each
unit so it can be used when training and sampling.

An example of data in a dataset would be the collection of letters from a character
neural network. Each letter would be mapped to a single input and output cell,
because the network uses groups of letters to predict groups of letters.

Another example might be inputs that are pixels coded by location and color. The outputs
could be categories of things that are in the photos, or that the photos represent.
*/
type Dataset struct {
	KeyToItem map[interface{}]PerceptionUnit
	cellToKey map[CellID]interface{}
}

/*
Trainer provides a way to run simulations on a neural network, then capture the results
and keep them, or re-strengthen the network.

The result of training is a new network which can be diffed onto an original network.
*/
type Trainer interface {
	PrepareData(*Network)
	OnTrained()
}

/*
Train executes the trainer's OnTrained method once complete.
*/
func Train(t Trainer, settings *TrainingSettings, network *Network) {
	t.PrepareData(network)

	settings.Data.cellToKey = make(map[CellID]interface{})

	for key, dataItem := range settings.Data.KeyToItem {
		// grow paths between all the inputs and outputs
		network.GrowPathBetween(dataItem.InputCell, dataItem.OutputCell, 10)
		// reverse the map
		settings.Data.cellToKey[dataItem.InputCell] = key
		settings.Data.cellToKey[dataItem.OutputCell] = key
		// prevent accidentally pruning the input/output cells
		network.Cells[dataItem.InputCell].Immortal = true
		network.Cells[dataItem.OutputCell].Immortal = true
	}

	// TODO: do actual training from shake.go

	t.OnTrained()
}
