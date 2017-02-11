package potential

import "math"

/*
laws.go is the collection of constants.

They are called "laws" because these are the ratios of a neural network's universe.
Sort of like Planck's constant or the speed of light, these laws define the universe
in which all neural network will live.

Any ratios or constants that keep the network reliably predictable across hardware
and training or sampling sessions should be in this file.
*/

// actualSynapseMin and actualSynapseMax helps make math less intensive if there is
// never a chance synapse addition will create an int8 overflow.
//
// DO NOT CHANGE THESE TWO.
// must be one outside the bounds, plus/minus the learning rate. otherwise
// the int8 will FLIP ITS PLUS or MINUS!!
const actualSynapseMin int16 = -127 + int16(synapseLearnRate)
const actualSynapseMax int16 = 126 - int16(synapseLearnRate)

/*
newSynapseMinMillivolts is the bottom range of how much a new synapse will
have for the `Millivolts` property.
*/
const newSynapseMinMillivolts int = -30

/*
newSynapseMaxMillivolts is the bottom range of how much a new synapse will
have for the `Millivolts` property.
*/
const newSynapseMaxMillivolts int = 30

/*
synapseLearnRate is how much a synapse should get bumped when it is being reinforced.

This is an absolute value because the synapse may be positive or negative, and this
value will be how much it is bumped away from zero.

When de-reinforcing or degrading a synapse, it will get reduced at 1/2 the distance
to zero, until it is 2, then it will become zero.
*/
const synapseLearnRate int8 = 2

/*
cellRestingVoltage comes from standard neuroscience Membrane Potential. This, and all voltages
in the lib, conveniently fit in tiny 8 bit integers.
*/
const cellRestingVoltage int8 = -10

/*
cellFireVoltageThreshold represents the millivolts where an action potential will result.
int16 is needed for comparisons.
*/
const cellFireVoltageThreshold int16 = 60

/*
synapseAPBoost is how much a synapse's ActivationHistory should be incremented extra when
its firing results in activation (strengthening the synapse)
*/
const synapseAPBoost uint = 1

// There are two factors that result in degrading a synapse:
/*
 */
const samplesBetweenMergingSessions = 16

/*
defaultSynapseMinFireThreshold represents the minimum we expect a synapse to fire between
pruning sessions (`samplesBetweenMergingSessions`), for this cell to get reinforced.
*/
const defaultSynapseMinFireThreshold = 8

/*
defaultNeuronSynapses is the number of random synapses a new neuron will get.
*/
const defaultNeuronSynapses = 4

/*
retrainNeuronsToGrow is the number of neurons to add when a single sample does not yield the expected output firing.
*/
const retrainNeuronsToGrow = 1

/*
retrainRandomSynapsesToGrow is the number of synapses to add when a single
session does not yield the expected output firing.
*/
const retrainRandomSynapsesToGrow = 0

/*
GrowPathExpectedMinimumSynapses represents the maximum allowed number of
synapses between an input and output cell in the network which get added
when an input cell fails to fire the output cell.

By default, it assumes each new synapse has a default mv value of they synapse learn rate (not
true for randomly grown synapses) and we will grow enough to essentially fire the cell if all
these synapses fire.
*/
const GrowPathExpectedMinimumSynapses = int((cellFireVoltageThreshold - int16(cellRestingVoltage)) / int16(synapseLearnRate))

/*
defaultNewGrownPathSynapse is the `Millivolts` value for a new synapse that is added during
`GrowPathBetween`. It should be enough to fire the next cell. Maybe with some wiggle room.

Its value should be considered in relation to `GrowPathExpectedMinimumSynapses`,
`cellRestingVoltage`, and `cellFireVoltageThreshold`.
*/
const defaultNewGrownPathSynapse int8 = 15

/*
ratioMaxHopsBetweenCellsDuringPathTrace is how many steps (synapses) are in between an
input an output cell before we forge a path up to `GrowPathExpectedMinimumSynapses` in
between them.
*/
func ratioMaxHopsBetweenCellsDuringPathTrace(network *Network) int {
	lenSyn := len(network.Synapses)
	lenCell := len(network.Cells)
	if lenCell == 0 {
		return 0 // prevents divide by zero panic
	}
	avgSynPerCell := float64(lenSyn / lenCell)
	// semi-hardcoded number of max hops. this was arbitrary.
	maxHops := int(math.Max(math.Min(avgSynPerCell, 20.0), 20))
	return maxHops
}
