package laws

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
const NewSynapseMinMillivolts int = -30

/*
NewSynapseMaxMillivolts is the bottom range of how much a new synapse will
have for the `Millivolts` property.
*/
const NewSynapseMaxMillivolts int = 30

/*
SynapseLearnRate is how much a synapse should get bumped when it is being reinforced.

This is an absolute value because the synapse may be positive or negative, and this
value will be how much it is bumped away from zero.

When de-reinforcing or degrading a synapse, it will get reduced at 1/2 the distance
to zero, until it is 2, then it will become zero.
*/
const SynapseLearnRate int16 = 2

/*
CellRestingVoltage is what a neuron gets reset to after it has fired.
*/
const CellRestingVoltage int16 = -10

/*
CellFireVoltageThreshold represents the millivolts where an action potential will result.
*/
const CellFireVoltageThreshold int = 100

/*
SynapseAPBoost is how much a synapse's ActivationHistory should be incremented extra when
its firing results in activation (strengthening the synapse)
*/
const SynapseAPBoost uint = 1

// There are two factors that result in degrading a synapse:

/*
SamplesBetweenMergingSessions is how many samples run on a network clone
instance before merging any synapse or neuron changes back to the master
network.
*/
const SamplesBetweenMergingSessions = 16

/*
DefaultSynapseMinFireThreshold represents the minimum we expect a synapse to fire between
pruning sessions (`SamplesBetweenMergingSessions`), for this cell to get reinforced.
*/
const DefaultSynapseMinFireThreshold = 8

/*
DefaultNeuronSynapses is the number of random synapses a new neuron will get.
*/
const DefaultNeuronSynapses = 8

/*
RetrainNeuronsToGrow is the number of neurons to add when a single sample does not
yield the expected output firing.
*/
const RetrainNeuronsToGrow = 1

/*
GrowPathExpectedMinimumSynapses represents the maximum allowed number of
synapses between an input and output cell in the network which get added
when an input cell fails to fire the output cell.

By default, it assumes each new synapse has a default mv value of they synapse learn rate (not
true for randomly grown synapses) and we will grow enough to essentially fire the cell if all
these synapses fire.
*/
const GrowPathExpectedMinimumSynapses = 2

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
const MaxPostFireSteps uint8 = 100

/*
FiringIterationsPerSample is how many times to fire an input cell.
Firing once may not cause much firing in the network. So firing
10 or 100+ times in a row will excite many pathways.
*/
const FiringIterationsPerSample int = 10

/*
PatternSimilarityLimit represents the percentage/ratio of
similarity between two firing patterns before one (or both?) of them
need to change.
*/
const PatternSimilarityLimit float64 = 0.7

/*
InitialCellCountPerVocabUnit is how many cells will represent a single
VocabUnit, to start off.
*/
const InitialCellCountPerVocabUnit int = 10

/*
NewCellDifferentiationCount is how many new cells to add to a vocab
unit when we find it is too similar to another vocab unit.
*/
const NewCellDifferentiationCount int = 4

/*
NoiseRatio is the percentage of cells to purposely fire as noise
during training (or sampling?).
*/
const NoiseRatio float64 = 0.3
