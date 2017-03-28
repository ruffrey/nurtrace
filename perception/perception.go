package perception

import (
	"github.com/ruffrey/nurtrace/potential"
)

/*
Perception describes a real world set of data that is mapped to neurons.


*/
type Perception interface {
	SetRawData(rawData []byte)
	SaveVocab(settings *potential.TrainingSettings, filename string) error
	LoadVocab(settings *potential.TrainingSettings, filename string) error
	PrepareData(settings *potential.TrainingSettings, network *potential.Network)
	SeedAndSample(settings *potential.TrainingSettings, seed string, network *potential.Network)
}
