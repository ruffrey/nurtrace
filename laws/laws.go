package laws

import "math"

/*
laws.go is the collection of constants.

They are called "laws" because these are the ratios of a neural network's universe.
Sort of like Planck's constant or the speed of light, these laws define the universe
in which all neural network will live.

Any ratios or constants that keep the network reliably predictable across hardware
and training or sampling sessions should be in this file.
*/

// ActualSynapseMin and ActualSynapseMax helps make math less intensive if there is
// never a chance synapse addition will create an overflow.
//
// DO NOT CHANGE THESE TWO.
// must be one outside the bounds, plus/minus the learning rate. otherwise
// the int16 will FLIP ITS PLUS or MINUS!!
const ActualSynapseMin int16 = -32767 + int16(SynapseLearnRate)

// ActualSynapseMax is documented above.
const ActualSynapseMax int16 = 32766 - int16(SynapseLearnRate)

/*
NewSynapseMinMillivolts is the bottom range of how much a new synapse will
have for the `Millivolts` property.
*/
const NewSynapseMinMillivolts int = -5

/*
NewSynapseMaxMillivolts is the bottom range of how much a new synapse will
have for the `Millivolts` property.
*/
const NewSynapseMaxMillivolts int = 5

/*
IdealCellSynapseBalance is a ratio of the number of cells per synapse.
There should always be more synapses than cells, so this value is between
0 and 1.

Examples:
- 10 synapses per cell would be a ratio of `0.1`.
- 20 synapses per cell would be a ratio of `.05`.
*/
const IdealCellSynapseBalance float64 = 0.005

var ComputedSynapsesPerCell int = int(math.Ceil(1 / IdealCellSynapseBalance))

/*
SynapseLearnRate is how much a synapse should get bumped when it is being reinforced.

This is an absolute value because the synapse may be positive or negative, and this
value will be how much it is bumped away from zero.

When de-reinforcing or degrading a synapse, it will get reduced at 1/2 the distance
to zero, until it is 2, then it will become zero.
*/
const SynapseLearnRate int16 = 2

/*
CellFireVoltageThreshold represents the millivolts where an action potential will result.
*/
const CellFireVoltageThreshold int = 1000

/*
CellRestingVoltage is what a neuron gets reset to after it has fired.
*/
const CellRestingVoltage int16 = -int16(CellFireVoltageThreshold / 6)

/*
MaxDepthFromInputToOutput is how far the path between an input cell and its
expected output cell we are willing to tolerate. If it is closer, OK. If it
is longer, we will forge a path that is this-number-of-cells deep.

Depth creates the opportunity for complexity and stateful decisions in the
network.
*/
const MaxDepthFromInputToOutput uint8 = 20

/*
MaxPostFireSteps is how long to keep firing the network while we collect
the pattern. A network will get seeded with some initial cells to fire,
then it will keep firing (stepping) while we record what gets fired.
If it doesn't fizzle out on its own, it will stop at MaxPostFireSteps.
*/
const MaxPostFireSteps int = 5

/*
FiringIterationsPerSample is how many times to fire an input cell.
Firing once may not cause much firing in the network. So firing
10 or 100+ times in a row will excite many pathways.
*/
const FiringIterationsPerSample int = 6

/*
PatternSimilarityLimit represents the percentage/ratio of
similarity between two firing patterns before one (or both?) of them
need to change.
*/
const PatternSimilarityLimit float64 = 0.8

/*
InitialCellCountPerInput is how many cells will represent a single
VocabUnit, to start off.
*/
const InitialCellCountPerInput int = 16

/*
InputCellDifferentiationCount is the number of cells to add in order to
make a set of inputs more different.
*/
const InputCellDifferentiationCount = 1

/*
NoiseRatio is the percentage of cells to purposely fire as noise
during training (or sampling?).
*/
const NoiseRatio float64 = 0.01

/*
TrainingMergeBackIteration is the point at which we reset a network during training.
*/
const TrainingMergeBackIteration = 10
