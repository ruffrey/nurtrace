package potential

import "sync"

// Backtracing is equivalent to backpropagation.

/*
backwardTraceFiringsGood traverses the trees backward, from output to input.

It returns the synapse IDs that helped fire the intended cell, and the "bad"
synapses which resulted in noise, and made the wrong output cells fire.

To find good synapses, follow excitatory cells from the expected output to the
input cell. Those excitatory cells may also have inhibitory synapses, but we
are not going to walk up them.
*/
func backwardTraceFirings(network *Network, fromOutput CellID, toInput CellID) (goodSynapses map[SynapseID]bool) {
	goodSynapses = make(map[SynapseID]bool)
	walkedCells := make(map[CellID]bool) // prevent walking forever in looping circuits

	var mux sync.Mutex
	var wg sync.WaitGroup
	ch := make(chan SynapseID)
	var walkBack func(cellID CellID)
	walkBack = func(cellID CellID) {
		if cellID == toInput {
			return
		}
		mux.Lock()
		if _, already := walkedCells[cellID]; already {
			mux.Unlock()
			return
		}
		walkedCells[cellID] = true
		mux.Unlock()

		wg.Add(1)
		go func() {
			// find all of synapses, then cells that could have fired this cell
			dendrites := network.GetCell(cellID).DendriteSynapses
			for synapseID := range dendrites {
				synapse := network.GetSyn(synapseID)
				axon := network.GetCell(synapse.FromNeuronAxon)
				// walk up the synapse to see if its cell was fired.
				// We want to keep walking up the excitatory cells.
				// We do not want to walk up inhibitory synapses,
				// but (for now at least?) save the inhibitory synapse.
				// TODO: necessary to save inhibitory synapse?
				// Is this even correct?
				if axon.WasFired {
					ch <- synapseID
					if synapse.Millivolts > 0 {
						walkBack(synapse.FromNeuronAxon)
					}
				}
			}
			wg.Done()
		}()
	}

	go func() {
		walkBack(fromOutput)
		wg.Wait()
		close(ch)
	}()

	for synapseID := range ch {
		goodSynapses[synapseID] = true
	}

	return goodSynapses
}

type badPair struct {
	noisyCell CellID
	goodCell  CellID
	voltage   int16
}

/*
backwardTraceNoiseAndInhibit returns the synapses whose pathways resulted in firing incorrect
output cells.

- step backward through the network from the unexpected output cells
- upon finding a happy path cell that fired, immediately add an inhibitory synapse
from it to the bad pathway cell.

NOT currently in use. But this was pretty hefty and complex to implement,
and it might be useful in the future.
*/
func backwardTraceNoiseAndInhibit(network *Network, inputCells map[CellID]bool, unexpectedOutputCells map[CellID]bool, goodSynapses map[SynapseID]bool) {
	var badPairs []badPair
	walkedCells := make(map[CellID]bool) // prevent walking forever in looping circuits

	var mux sync.Mutex
	var wg sync.WaitGroup
	var ch chan badPair
	var walkBack func(cellID CellID)
	// We will start at an unexpected output cell and traverse backward
	// up its dendrites. Goal being to hit a good synapse which will
	// tell us to stop and use that good synapse's dendrite cell to inhibit
	// the cellID.
	walkBack = func(cellID CellID) {
		if _, isInputCell := inputCells[cellID]; isInputCell {
			return
		}
		mux.Lock()
		if _, already := walkedCells[cellID]; already {
			mux.Unlock()
			return
		}
		walkedCells[cellID] = true
		mux.Unlock()

		wg.Add(1)
		go func() {
			// find all of synapses, then cells that could have fired this cell
			dendrites := network.GetCell(cellID).DendriteSynapses
			for synapseID := range dendrites {
				// If this bad cell was fired by a good synapse,
				// use the good synapse's dendrite cell to inhibit
				// this bad cell.
				synapse := network.GetSyn(synapseID)
				excitatory := synapse.Millivolts > 0
				if _, isGood := goodSynapses[synapseID]; isGood && excitatory {
					bp := badPair{
						goodCell:  synapse.FromNeuronAxon,
						noisyCell: cellID,
						voltage:   synapse.Millivolts,
					}
					ch <- bp
					continue
				}

				if !excitatory {
					continue
				}
				axon := network.GetCell(synapse.FromNeuronAxon)
				if !axon.WasFired {
					continue
				}
				walkBack(synapse.FromNeuronAxon)
			}
			wg.Done()
		}()
	}

	for cellID := range unexpectedOutputCells {
		ch = make(chan badPair)
		go func() {
			walkBack(cellID)
			wg.Wait()
			close(ch)
		}()

		for bp := range ch {
			badPairs = append(badPairs, bp)
		}
	}

	// add the synapses afterward to prevent changing the network while
	// it is still being traversed.
	for _, bp := range badPairs {
		addInhibitorSynapse(network, bp.noisyCell, bp.goodCell, bp.voltage)
	}
}

func addInhibitorSynapse(network *Network, noisyCell CellID, goodAxonFutureInhibitor CellID, positiveVoltage int16) SynapseID {
	// This inhibitor is a new synapse that will counteract
	// the "noisy" synapse which contributed to the wrong cell firing.
	inhibitor := network.linkCells(goodAxonFutureInhibitor, noisyCell)
	inhibitor.Millivolts = -positiveVoltage
	// log.Println("added inhibitor", inhibitor)
	return inhibitor.ID
}
