package potential

import "sync"

/*
backwardTraceFiringsGood traverses the trees backward, from output to input.

It returns the synapse IDs that helped fire the intended cell, and the "bad"
synapses which resulted in noise, and made the wrong output cells fire.

To find good synapses, follow excitatory cells from the expected output to the
input cell.
*/
func backwardTraceFirings(network *Network, fromOutput CellID, toInput CellID) (goodSynapses map[SynapseID]bool) {
	goodSynapses = make(map[SynapseID]bool)

	var wg sync.WaitGroup
	var mux sync.Mutex
	var walkBack func(cellID CellID)
	walkBack = func(cellID CellID) {
		if cellID == toInput {
			return
		}

		wg.Add(1)
		go func() {
			// find all of synapses, then cells that could have fired this cell
			for synapseID := range network.Cells[cellID].DendriteSynapses {
				synapse := network.Synapses[synapseID]
				axon := network.Cells[synapse.FromNeuronAxon]
				isExcitatory := synapse.Millivolts > 0
				// walk up the synapse to see if its cell was fired
				// and if it was excitatory. We want to keep walking
				// up the excitatory cells.
				if axon.WasFired && isExcitatory {
					mux.Lock()
					goodSynapses[synapseID] = true
					mux.Unlock()
					walkBack(synapse.FromNeuronAxon)
				}
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
	}()

	walkBack(fromOutput)

	return goodSynapses
}

/*
backwardTraceNoise returns the synapses whose pathways resulted in firing incorrect
output cells.

- follow the cells and synapses backward from the unexpected output cell to the
original input cell.
- ignore cells that didn't fire
- ignore cells that were on the happy path
- ignore synapses that weren't excitatory
*/
func backwardTraceNoise(network *Network, inputCell CellID, unexpectedOutputCells []CellID, goodSynapses map[SynapseID]bool) (badSynapses map[SynapseID]bool) {
	badSynapses = make(map[SynapseID]bool)

	var wg sync.WaitGroup
	var mux sync.Mutex
	var walkBack func(cellID CellID)
	walkBack = func(cellID CellID) {
		if cellID == inputCell {
			return
		}

		wg.Add(1)
		go func() {
			// find all of synapses, then cells that could have fired this cell
			for synapseID := range network.Cells[cellID].DendriteSynapses {
				// the synapse is already known to be on the good path
				if _, isGood := goodSynapses[synapseID]; isGood {
					continue
				}
				synapse := network.Synapses[synapseID]
				notExcitatory := synapse.Millivolts < 1
				if notExcitatory {
					continue
				}
				axon := network.Cells[synapse.FromNeuronAxon]
				if !axon.WasFired {
					continue
				}
				// so this isn't so much paths as all the bad synapses
				// finding the one-way paths is trickier, and an optimization
				// problem
				mux.Lock()
				badSynapses[synapseID] = true
				mux.Unlock()
				walkBack(synapse.FromNeuronAxon)
			}
			wg.Done()
		}()
	}

	for _, cellID := range unexpectedOutputCells {
		go walkBack(cellID)
		wg.Wait()
	}

	return badSynapses
}

/*
applyBacktrace adds a synapse to inhibit the any bad path that produced noise, i.e.
it resulted in the wrong output cell firing and was not in the path to the right
output cell.

It  reinforces the good path synapses, too.
*/
func applyBacktrace(network *Network, goodSynapses map[SynapseID]bool, badPathEntrySynapses map[SynapseID]bool) {
	for synapseID := range goodSynapses {
		network.Synapses[synapseID].reinforce()
	}
	for noisySynapseID := range badPathEntrySynapses {
		inhibitor := NewSynapse(network)
		inhibitor.Millivolts = -network.Synapses[noisySynapseID].Millivolts
		outletCellID := network.Synapses[noisySynapseID].ToNeuronDendrite

		network.Cells[outletCellID].addDendrite(inhibitor.ID)
	}
}
