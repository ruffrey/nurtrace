package potential

import (
	"fmt"
	"math/rand"
	"sync"
)

/*
backwardTraceFiringsGood traverses the trees backward, from output to input.

It returns the synapse IDs that helped fire the intended cell, and the "bad"
synapses which resulted in noise, and made the wrong output cells fire.

To find good synapses, follow excitatory cells from the expected output to the
input cell.
*/
func backwardTraceFirings(network *Network, fromOutput CellID, toInput CellID) (goodSynapses map[SynapseID]bool) {
	r := rand.Uint32()
	fmt.Println("good start", r)
	goodSynapses = make(map[SynapseID]bool)

	var wg sync.WaitGroup
	ch := make(chan SynapseID)
	var walkBack func(cellID CellID)
	walkBack = func(cellID CellID) {
		if cellID == toInput {
			return
		}

		wg.Add(1)
		go func() {
			// find all of synapses, then cells that could have fired this cell
			network.cellMux.Lock()
			dendrites := network.Cells[cellID].DendriteSynapses
			network.cellMux.Unlock()
			for synapseID := range dendrites {
				network.synMux.Lock()
				synapse := network.Synapses[synapseID]
				network.synMux.Unlock()
				network.cellMux.Lock()
				axon := network.Cells[synapse.FromNeuronAxon]
				network.cellMux.Unlock()
				isExcitatory := synapse.Millivolts > 0
				// walk up the synapse to see if its cell was fired
				// and if it was excitatory. We want to keep walking
				// up the excitatory cells.
				if axon.WasFired && isExcitatory {
					ch <- synapseID
					walkBack(synapse.FromNeuronAxon)
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

	fmt.Println("good end ", r, ", goodSynapses=", len(goodSynapses),
		"synapses=", len(network.Synapses))
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
func backwardTraceNoise(network *Network, inputCells map[CellID]bool, unexpectedOutputCells map[CellID]bool, goodSynapses map[SynapseID]bool) (badSynapses map[SynapseID]bool) {
	r := rand.Uint32()
	fmt.Println("bad  start", r)
	badSynapses = make(map[SynapseID]bool)

	var wg sync.WaitGroup
	var ch chan SynapseID
	var walkBack func(cellID CellID)
	walkBack = func(cellID CellID) {
		if _, isInputCell := inputCells[cellID]; isInputCell {
			return
		}

		wg.Add(1)
		go func() {
			// find all of synapses, then cells that could have fired this cell
			network.cellMux.Lock()
			dendrites := network.Cells[cellID].DendriteSynapses
			network.cellMux.Unlock()
			for synapseID := range dendrites {
				// the synapse is already known to be on the good path
				if _, isGood := goodSynapses[synapseID]; isGood {
					continue
				}
				network.synMux.Lock()
				synapse := network.Synapses[synapseID]
				network.synMux.Unlock()
				notExcitatory := synapse.Millivolts < 1
				if notExcitatory {
					continue
				}
				network.cellMux.Lock()
				axon := network.Cells[synapse.FromNeuronAxon]
				network.cellMux.Unlock()
				if !axon.WasFired {
					continue
				}
				// This isn't so much "paths" as just all the bad synapses.
				// Finding the one-way paths is trickier, and an optimization
				// problem.
				ch <- synapseID
				walkBack(synapse.FromNeuronAxon)
			}
			wg.Done()
		}()
	}

	for cellID := range unexpectedOutputCells {
		ch = make(chan SynapseID)
		go func() {
			walkBack(cellID)
			wg.Wait()
			close(ch)
		}()

		for synapseID := range ch {
			badSynapses[synapseID] = true
		}
	}

	fmt.Println("bad  end ", r, "badSynapses=", len(badSynapses),
		"synapses=", len(network.Synapses))
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
