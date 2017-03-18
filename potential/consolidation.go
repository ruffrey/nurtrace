package potential

import (
	"bleh/laws"
	"math"
	"strconv"
)

func synapseSignature(synapse *Synapse) string {
	return strconv.Itoa(int(synapse.FromNeuronAxon)) + "-" + strconv.Itoa(int(synapse.ToNeuronDendrite))
}

type dupeSynapses []SynapseID

/*
findDupeSynapses returns any group of synapses from a network where the
pre and post synaptic neurons are the same. Duplicate means they have
the same inputs and outputs, and theoretically can have their weights
combined and become one synapse.
*/
func findDupeSynapses(network *Network) map[string]dupeSynapses {
	synapsesByDendrite := make(map[string]dupeSynapses)
	dupes := make(map[string]dupeSynapses)

	for _, synapse := range network.Synapses {
		sig := synapseSignature(synapse)
		synapsesByDendrite[sig] = append(synapsesByDendrite[sig], synapse.ID)
	}

	for sig, similarSynapses := range synapsesByDendrite {
		if len(similarSynapses) > 1 {
			dupes[sig] = similarSynapses
		}
	}

	return dupes
}

const actualSynapseMaxFloat64 = float64(laws.ActualSynapseMax)

/*
dedupeSynapses receives a list of synapses that are known to have
the same inputs and outputs, removing as many as possible
*/
func dedupeSynapses(synapses dupeSynapses, network *Network) {
	var sum float64
	var keepSynapses []SynapseID
	var removeSynapses []SynapseID
	dupeSynapsesTotal := len(synapses)

	for _, synapseID := range synapses {
		sum += float64(network.getSyn(synapseID).Millivolts) // unlikely to overflow, but may
	}

	keepTotal := int(math.Ceil(math.Abs(sum) / actualSynapseMaxFloat64))
	isPositive := sum >= 0
	var max int16
	if isPositive {
		max = laws.ActualSynapseMax
	} else {
		max = laws.ActualSynapseMin
	}

	if keepTotal == dupeSynapsesTotal {
		return
	}

	keepSynapses = synapses[0:keepTotal]
	removeSynapses = synapses[keepTotal:]
	lastKeepNewMillivolts := int16((dupeSynapsesTotal * int(max)) - (keepTotal * int(max)))

	lastIndex := len(keepSynapses) - 1
	for i, synapseID := range keepSynapses {
		isLast := lastIndex == i
		if isLast {
			network.getSyn(synapseID).Millivolts = lastKeepNewMillivolts
		} else {
			network.getSyn(synapseID).Millivolts = max
		}
	}
	for _, synapseID := range removeSynapses {
		network.PruneSynapse(synapseID)
	}
}