package potential

import (
	"errors"
	"fmt"
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

	return badSynapses
}

/*
applyBacktrace inhibits the "bad" paths that produced noise, i.e.
it 1) resulted in the wrong output cell firing and 2) was not in
the path to the right output cell.

It tries to just bolser an existing synapse from a good path synapse,
but will create a new one if necessary.

It reinforces the good path synapses, too.
*/
func applyBacktrace(network *Network, inputCells map[CellID]bool, goodSynapses map[SynapseID]bool, badPathEntrySynapses map[SynapseID]bool) {
	lenGoodSynapses := len(goodSynapses)
	goodAxons := make(map[CellID]bool)                // might need these later
	dendriteToSynapseID := make(map[CellID]SynapseID) // might need later, ok if multiple overwrite

	for synapseID := range goodSynapses {
		synapse := network.Synapses[synapseID]
		synapse.reinforce()
		goodAxons[synapse.FromNeuronAxon] = true
		dendriteToSynapseID[synapse.ToNeuronDendrite] = synapse.ID
	}
	fmt.Println("applyBacktrace\n  badSynapses", len(badPathEntrySynapses), "\n  goodSynapses", lenGoodSynapses, "\n  goodAxons", len(goodAxons), "\n  dendriteToSynapseID", dendriteToSynapseID)

	for noisySynapseID := range badPathEntrySynapses {
		noisySynapse := network.Synapses[noisySynapseID]
		fmt.Println("  noisySynapse.FromNeuronAxon", noisySynapse.FromNeuronAxon)
		// Try to see if we can reuse a synapse, before making a new one.
		// What we need is an existing good synapse that already inhibits
		// the cell that is the noisy synapse's dendrite.
		synapseID, err := findSynapseInhibitingCell(network, noisySynapse.FromNeuronAxon, dendriteToSynapseID)
		fmt.Println("  err", err)
		if err == nil {
			// found an existing synapse
			fmt.Println("  reusing existing synapse", synapseID)
			reinforceByAmount(network.Synapses[synapseID], -noisySynapse.Millivolts)
			continue
		}
		// fmt.Println("  making new synapse")
		// create a new synapse and have a random good synapse axon
		// cell fire the inhibitor
		var randCellID CellID
		if lenGoodSynapses > 0 {
			randCellID = randCell(goodAxons)
			// fmt.Println("  using good axon", randCellID)
		} else {
			randCellID = randCell(inputCells)
			// fmt.Println("  using input cell", randCellID)
		}
		addInhibitorSynapse(network, noisySynapse, randCellID)
	}
}

var errNoSynapseExistingFound = errors.New("no existing synapse") // internal only

func findSynapseInhibitingCell(network *Network, cellNeedingInhibit CellID, dendriteToSynapseReverseMap map[CellID]SynapseID) (SynapseID, error) {
	if sid, exists := dendriteToSynapseReverseMap[cellNeedingInhibit]; exists {
		synapse := network.Synapses[sid]
		if synapse.Millivolts >= 0 {
			return 0, errNoSynapseExistingFound // do not want excitatory ones
		}
		// does this synapse fire the cell we need?
		if synapse.ToNeuronDendrite == cellNeedingInhibit {
			return sid, nil
		}
	}

	return 0, errNoSynapseExistingFound
}

func addInhibitorSynapse(network *Network, noisySynapse *Synapse, axon CellID) {
	// This inhibitor is a new synapse that will counteract
	// the "noisy" synapse which contributed to the wrong cell firing.
	inhibitor := NewSynapse(network)
	inhibitor.Millivolts = -noisySynapse.Millivolts
	unwantedOutputCell := network.Cells[noisySynapse.FromNeuronAxon]

	unwantedOutputCell.addDendrite(inhibitor.ID)

	network.Cells[axon].addAxon(inhibitor.ID)
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
