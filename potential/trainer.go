package potential

import "time"

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
	Data    Dataset
	// List of arrays of cells to fire for training.
	TrainingSamples [][]CellID
	// Solutions should correspond to the samples and will be matched on index.
	TrainingSolutions [][]CellID
}

/*
DataItem is the smallest and core unit of a Dataset.
*/
type DataItem struct {
	Value      interface{}
	InputCell  CellID
	OutputCell CellID
}

/*
Dataset is
*/
type Dataset struct {
	KeyToItem map[interface{}]DataItem
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
func Train(t *Trainer, settings *TrainingSettings, network *Network) {
	t.PrepareData(network)

	if len(settings.TrainingSamples) != len(settings.TrainingSolutions) {
		panic("TrainingSamples and TrainingSolutions should be paired on indexes and therefore equal in length.")
	}

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
