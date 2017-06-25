package potential

import (
	"sync"

	"github.com/ruffrey/nurtrace/laws"
	"log"
)

// All methods on a network that relate to growing are here.

/*
Grow is a general growth that encompasses all growth methods.
*/
func (network *Network) Grow(neuronsToAdd, synapsesPerNewNeuron, synapsesToAdd int) {
	someSynapses := synapsesPerNewNeuron / 4
	network.GrowRandomNeurons(neuronsToAdd, someSynapses)
	network.GrowRandomSynapses(synapsesToAdd)
	for i := 0; i < neuronsToAdd; i++ {
		if i % 25 == 0 {
			log.Println("growing deep synapses, progress=", i, "/", neuronsToAdd,
				"totalSynapses=", len(network.Synapses))
		}
		from := network.RandomCellKey()
		var to CellID
		for {
			to = network.RandomCellKey()
			if to != from {
				break
			}
		}
		network.GrowPathBetween(from, to, laws.ComputedSynapsesPerCell)
	}
}

/*
GrowPathBetween will make a path between two neurons.

The desired goal is for there to be a path, or series of paths, from startCell to endCell.

1. We start in at the startCell and traverse its axons.
2. At each layer, we check to see if any of the start axons are connected to the
end dendrites.
3. We count the connected synapses as we go.
4. Upon reaching minSynapses, or maxDepth, or nowhere, we stop.
	- if reached minSynapses, just really stop and return 0 synapses added
	- if reached maxDepth, continue.
5. We need to add enough synapses to reach minSynapses, such that these are the minimum
number of synapses that connect from startCell's tree to endCell.

After `maxDepth`, if there are not minSynapses, we create synapses at that layer.
*/
func (network *Network) GrowPathBetween(startCell, endCell CellID, minSynapses int) (synapsesToEnd map[SynapseID]bool, synapsesAdded map[SynapseID]bool) {
	// these are the synapses we found that are on the path from the startCell,
	// and attach directly to an endCell at the dendrite
	synapsesToEnd = make(map[SynapseID]bool)
	var synapsesToEndMux sync.Mutex
	// any new synapses we create if there are not enough in the network that attach
	// to the end cell
	synapsesAdded = make(map[SynapseID]bool)

	const maxDepth = laws.MaxDepthFromInputToOutput
	var depthReached uint8

	mux := sync.Mutex{}
	var wg sync.WaitGroup
	ch := make(chan SynapseID)
	alreadyWalked := make(map[CellID]bool)
	lastDepth := make(map[CellID]bool)

	// walk traverses the axons and see if any synapse leads to the end cell.
	// hops is the layer we are on, *copied* on pass.
	// this is a fan-out kind of traversal through a tree of cells and synapses.
	// Note: need to fully declare it before assigning it, apparently because the
	// runtime needs this to compile a recursive function.
	var walk func(cellID CellID, currentDepth uint8)
	walk = func(cellID CellID, currentDepth uint8) {
		mux.Lock()
		synapsesToEndMux.Lock()
		if currentDepth > depthReached {
			depthReached = currentDepth
			lastDepth = make(map[CellID]bool)
		}
		lastDepth[cellID] = true
		totalSynapsesFound := len(synapsesToEnd)
		synapsesToEndMux.Unlock()
		if currentDepth > maxDepth || totalSynapsesFound >= minSynapses {
			mux.Unlock()
			return
		}
		if _, already := alreadyWalked[cellID]; already {
			mux.Unlock()
			return
		}
		alreadyWalked[cellID] = true
		mux.Unlock()

		wg.Add(1)
		// look at the next cells in the axon chain from this one, to see
		// if any are the endCell then send it down the channel.
		go func() {
			axonSynapses := network.GetCell(cellID).AxonSynapses
			for axonSynapseID := range axonSynapses {
				s := network.GetSyn(axonSynapseID)
				receiverCellID := s.ToNeuronDendrite

				// It only counts if it is excitatory.
				// We also walk the axons of this cell.
				if s.Millivolts > 0 {
					if receiverCellID == endCell {
						ch <- axonSynapseID
					}
					walk(receiverCellID, currentDepth+1)
				}
			}
			wg.Done()
		}()
	}

	// receive the connected synapses.
	// channel is closed upstream when we reach the end
	go func() {
		walk(startCell, 0)
		wg.Wait()
		close(ch)
	}()

	for synapseToOutputCell := range ch {
		// log.Println("synapseToOutputCell", synapseToOutputCell)
		synapsesToEndMux.Lock()
		synapsesToEnd[synapseToOutputCell] = true
		synapsesToEndMux.Unlock()
	}

	needSynapses := minSynapses - len(synapsesToEnd) // out of multithreading now; no mutex
	if needSynapses > 0 {
		lastCell := randCellFromMap(lastDepth)
		alt := true
		for i := 0; i < needSynapses-1; i++ {
			var intermediary CellID
			if alt {
				intermediary = randCellFromMap(alreadyWalked)
			} else {
				intermediary = network.RandomCellKey()
			}
			alt = !alt

			newLinkingSynapse := network.linkCells(lastCell, intermediary)

			synapsesAdded[newLinkingSynapse.ID] = true

			newLinkingSynapse.Millivolts = int16(laws.CellFireVoltageThreshold)
			lastCell = intermediary
		}

		newLinkingSynapse := network.linkCells(lastCell, endCell)
		synapsesAdded[newLinkingSynapse.ID] = true
		newLinkingSynapse.Millivolts = int16(laws.CellFireVoltageThreshold)
	}

	// Reinforce the path between expected input and output.
	// Since goodSynapses looks for cells that actually fired,
	// it may take more than one entire itration for this next
	// round of code to do anything. Because the new pathways
	// generated in the block above did not necessarily fire.
	goodSynapses := backwardTraceFirings(network, endCell, startCell)
	for synapseID := range goodSynapses {
		network.Synapses[synapseID].reinforce()
	}

	return synapsesToEnd, synapsesAdded
}

/*
GrowRandomNeurons will randomly add neurons with the default number of synapses to the network.
*/
func (network *Network) GrowRandomNeurons(neuronsToAdd, synapsesPerNeuron int) {
	// log.Println("  adding neurons =", neuronsToAdd)
	// Now - all the new neurons are added first with no synapses. If synapses were added at
	// create time, the newer neurons would end up with far fewer connections to the following
	// newer neurons.
	var addedNeurons []*Cell
	for i := 0; i < neuronsToAdd; i++ {
		cell := NewCell(network)
		addedNeurons = append(addedNeurons, cell)
	}

	// Now we add the default number of synapses to our new neurons, with random other neurons.
	// Create the synapse, then choose a random cell from the network. This cell
	// will be the sender to the random cell.
	for _, cell := range addedNeurons {
		for i := 0; i < synapsesPerNeuron; {
			ix := network.RandomCellKey()
			otherCell := network.Cells[ix]
			if cell.ID == otherCell.ID {
				// try again
				continue
			}

			network.linkCells(cell.ID, otherCell.ID)
			i++
		}
	}
}

/*
GrowRandomSynapses adds the specified number of synapses haphazardly to the network.
*/
func (network *Network) GrowRandomSynapses(synapsesToAdd int) {
	for i := 0; i < synapsesToAdd; {
		senderIx := network.RandomCellKey()
		receiverIx := network.RandomCellKey()
		sender := network.Cells[senderIx]
		receiver := network.Cells[receiverIx]
		// Thy cell shannot activate thyself
		if sender.ID == receiver.ID {
			continue
		}

		network.linkCells(sender.ID, receiver.ID)
		i++
	}
}
