package perception

import "bleh/potential"

/*
Perception describes a real world set of data that is mapped to neurons.


*/
type Perception interface {
	GetSettings() *potential.TrainingSettings
	SetRawData(rawData []byte)
	SaveVocab(filename string) error
	LoadVocab(filename string) error
	PrepareData(network *potential.Network)
	SeedAndSample(seed string, network *potential.Network)
}
