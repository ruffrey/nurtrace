package potential

import (
	"fmt"
	"sync"
)

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
			dendrites := network.getCell(cellID).DendriteSynapses
			for synapseID := range dendrites {
				synapse := network.getSyn(synapseID)
				axon := network.getCell(synapse.FromNeuronAxon)
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

/*
backwardTraceNoise returns the synapses whose pathways resulted in firing incorrect
output cells.

- step backward through the network from the unexpected output cells
- find cells that fired and that were on the happy path
- upon finding a cell that fired that was on the happy path, stop stepping
and save that as the beginning of the noisy path to be inhibited later.

TODO: ^^ make this how it works. It doesn't work like this now.
*/
func backwardTraceNoise(network *Network, inputCells map[CellID]bool, unexpectedOutputCells map[CellID]bool, goodSynapses map[SynapseID]bool) (badSynapses map[SynapseID]bool) {
	badSynapses = make(map[SynapseID]bool)
	walkedCells := make(map[CellID]bool) // prevent walking forever in looping circuits

	var mux sync.Mutex
	var wg sync.WaitGroup
	var ch chan SynapseID
	var walkBack func(cellID CellID)
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
			dendrites := network.getCell(cellID).DendriteSynapses
			for synapseID := range dendrites {
				// the synapse is already known to be on the good path
				if _, isGood := goodSynapses[synapseID]; isGood {
					continue
				}
				synapse := network.getSyn(synapseID)
				notExcitatory := synapse.Millivolts < 1
				if notExcitatory {
					continue
				}
				axon := network.getCell(synapse.FromNeuronAxon)
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

	return badSynapses
}

/*
applyBacktrace inhibits the "bad" paths that produced noise, i.e.
it 1) resulted in the wrong output cell firing and 2) was not in
the path to the right output cell.

After listing all the bad cells that fired, we step through the network
from each goodSynapse, and upon finding the first step with bad synapses,
add inhibitory synapses on the cells *before* the bad synapses to they
don't fire. Then stop.

It reinforces the good path synapses, too.
*/
func applyBacktrace(network *Network, inputCells map[CellID]bool, goodSynapses map[SynapseID]bool, badPathEntrySynapses map[SynapseID]bool) {
	cellsBeingInhibitedByGoodSynapses := make(map[CellID]SynapseID) // might need later, ok if multiple overwrite
	goodAxons := make(map[CellID]bool)                              // might need these later
	// reinforce all good synapses. These are synapses that fire when one of the
	// input cells fires.
	for synapseID := range goodSynapses {
		synapse := network.Synapses[synapseID]
		synapse.reinforce()
		goodAxons[synapse.FromNeuronAxon] = true
		if synapse.Millivolts < 0 {
			cellsBeingInhibitedByGoodSynapses[synapse.ToNeuronDendrite] = synapse.ID
		}
	}

	// Stop on the first step that has bad axons. Inhibit all those axons
	// by reinforcing either an existing good path axon, or creating a new
	// synapse that inhibits the bad axon.
	// TODO: concurrency
	firstPathBadSynapses := make(map[SynapseID]bool)
	walkedCells := make(map[CellID]bool)
	var nextCells map[CellID]bool
	nextCells = inputCells
	// Later, if we need to create a new synapse, we want to make one at
	// the same step as the bad synapse, or one step before to give the
	// best chance for the good cell to inhibit the bad one.
	badSynapseToStepIndex := make(map[SynapseID]int)
	stepIndexToGoodSynapses := make(map[int]map[SynapseID]bool)
	stepIndex := 0

	for {
		if len(nextCells) == 0 {
			break
		}

		stepIndex++
		stepIndexToGoodSynapses[stepIndex] = make(map[SynapseID]bool)

		shouldStop := false

		for cellID := range nextCells {
			if _, already := walkedCells[cellID]; already {
				continue
			}
			walkedCells[cellID] = true

			cell := network.Cells[cellID]
			for s := range cell.AxonSynapses {
				// did this cell fire a bad synapse?
				if _, firedBadSynapse := badPathEntrySynapses[s]; firedBadSynapse {
					badSynapseToStepIndex[s] = stepIndex
					firstPathBadSynapses[s] = true
					shouldStop = true
					continue
				}
				// populate the next cells
				if _, firedGoodSynapse := goodSynapses[s]; firedGoodSynapse {
					synapse := network.Synapses[s]
					nextCells[synapse.ToNeuronDendrite] = true
					stepIndexToGoodSynapses[stepIndex][s] = true
				}
			}
		}

		if shouldStop {
			break
		}
		nextCells = make(map[CellID]bool)
	}

	// if len(firstPathBadSynapses) != len(badPathEntrySynapses) {
	// 	fmt.Println("firstPathBadSynapses=", len(firstPathBadSynapses), "badPathEntrySynapses", len(badPathEntrySynapses))
	// }

	// now that we know the beginning paths to bad synapses,
	// and they will fire regardless, we are going to inhibit each
	// one's dendrite (the cell it fires).

	for noisySynapseID := range firstPathBadSynapses {
		noisySynapse := network.Synapses[noisySynapseID]
		badSynapseStep := badSynapseToStepIndex[noisySynapseID]
		// Try to see if we can reuse a good synapse, before making a new one.
		// What we need is an existing good synapse that already inhibits
		// the cell that is the noisy synapse's dendrite.
		if inhibitorySynapseID, exists := cellsBeingInhibitedByGoodSynapses[noisySynapse.ToNeuronDendrite]; exists {
			// found an existing synapse
			if stepIndexToGoodSynapses[badSynapseStep][inhibitorySynapseID] {
				fmt.Println("  reusing existing synapse", inhibitorySynapseID)
				reinforceByAmount(network.Synapses[inhibitorySynapseID], noisySynapse.Millivolts)
				continue
			}
		}
		// create a new synapse and have a random good synapse axon at the same
		// step fire it.
		var goodCellToInhibitNoise CellID
		sstep := stepIndexToGoodSynapses[badSynapseStep]
		if len(sstep) > 0 {
			// Originally this loaded all axons in a list and selected one
			// at random. That was super, super slow, on the order of many
			// seconds (18s out of 60s). Since Go guarantees random order
			// of map keys during a range operation, we can just select one.
			for sid := range sstep {
				s := network.Synapses[sid]
				goodCellToInhibitNoise = s.FromNeuronAxon
				break // yes, break on first one since it was random
			}
		} else {
			goodCellToInhibitNoise = randCell(inputCells)
			// fmt.Println("Using cell from input", goodCellToInhibitNoise)
		}
		addInhibitorSynapse(network, noisySynapse, goodCellToInhibitNoise)
	}
}

func addInhibitorSynapse(network *Network, noisySynapse *Synapse, goodAxonFutureInhibitor CellID) SynapseID {
	// This inhibitor is a new synapse that will counteract
	// the "noisy" synapse which contributed to the wrong cell firing.
	inhibitor := network.linkCells(goodAxonFutureInhibitor, noisySynapse.ToNeuronDendrite)
	inhibitor.Millivolts = -noisySynapse.Millivolts

	return inhibitor.ID
}

func randCell(cellMap map[CellID]bool) (randCellID CellID) {
	iterate := randomIntBetween(0, len(cellMap)-1)
	i := 0
	for k := range cellMap {
		if i == iterate {
			randCellID = CellID(k)
			break
		}
		i++
	}
	if randCellID == CellID(0) {
		panic("Should never get cell ID 0")
	}
	return randCellID
}
