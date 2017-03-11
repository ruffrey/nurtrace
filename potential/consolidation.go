package potential

import (
	"strconv"
)

func synapseSignature(synapse *Synapse) string {
	return strconv.Itoa(int(synapse.FromNeuronAxon)) + "-" + strconv.Itoa(int(synapse.ToNeuronDendrite))
}

// findDupeSynapses returns any group of synapses where the
// pre and post synaptic neurons are
func findDupeSynapses(network *Network) map[string][]SynapseID {
	synapsesByDendrite := make(map[string][]SynapseID)
	dupes := make(map[string][]SynapseID)

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

func dedupeSynapses(synapes []SynapseID, network *Network) {

}
