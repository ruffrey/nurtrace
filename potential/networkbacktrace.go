package potential

/*
backwardTraceFiringsGood traverses the trees backward, from output to input.

Synapses that were in the firing path are marked as `goodPath:true`.
*/
func (network *Network) backwardTraceFiringsGood(fromOutput CellID, toInput CellID) {

}

/*
backwardTraceFiringsBad traverses the trees backward, from output to input.

Synapses that were in the firing path are marked as `badpath:true`.
*/
func (network *Network) backwardTraceFiringsBad(fromOutput CellID, toInput CellID) {

}

/*
applyBacktrace flips the first divergence to bad paths to be negative, and reinforces
the good path synapses.
*/
func (network *Network) applyBacktrace() {
	for _, synapse := range network.Synapses {
		if synapse.badPath {
			synapse.flip()
			continue
		}
		if synapse.goodPath {
			synapse.reinforce()
		}
	}
}
