package potential

import (
	"fmt"

	"github.com/ruffrey/nurtrace/laws"
)

/*
CellID should be unique for all cells in a network.
*/
type CellID uint32

/*
Cell holds voltage, receives input from Dendrites, and upon reaching the activation voltage,
fires an action potential cycle and its axon synapses push voltage to the dendrites it connects
to.

Maps are used for DendriteSynapses because they are easier to use for the use cases in this
library than slices. They are always the value true. Being false is not advised, but that is
undefined behavior. We can't have empty values in a map, so a bool is one byte.

Surprisingly, a map with integer keys and single byte values is the same size as an slice with
integer values. So it makes no difference at scale whether it is a map or array. If anything,
map lookups would be faster.

Using maps where one might expect arrays or slices (because it really is just a list) might
be a source of confusion.
*/
type Cell struct {
	ID CellID
	/*
	   Immortal means this cell cannot be pruned. It should only be by perceptors and
	   receptors.
	*/
	Immortal   bool
	Network    *Network `json:"-"` // skip circular reference in JSON
	Voltage    int16    // unnecessary to recreate cell
	activating bool     // unnecessary to recreate cell
	/*
	  DendriteSynapses are this cell's inputs. They are IDs of synapses.
	*/
	DendriteSynapses map[SynapseID]bool
	/*
	  DendriteSynapses are this cell's outputs. They are IDs of synapses. When it fires,
	  these synapses will be triggered.
	*/
	AxonSynapses map[SynapseID]bool
	/*
	  WasFired is used during training to know if this cell fired during the session.
	  TODO: may no longer be necessary because 1) activating property might handle
	  this, and 2) only used during backtracing which is dead code at the time of
	  writing.
	*/
	WasFired bool
	/*
	   OnFired is useful when this cell is an output cell, or during training. In simple
	   tests, callback are hugely more performant than channels.
	*/
	OnFired []func(CellID) `json:"-"`
	/*
		Purely informational, for use when testing or debugging.
	*/
	Tag string
}

/*
NewCell instantiates a Cell and returns a pointer to it.

Sets the network pointer to the supplied network.
*/
func NewCell(network *Network) *Cell {
	network.cellMux.Lock()
	c := Cell{
		ID:               CellID(len(network.Cells)),
		Network:          network,
		Immortal:         false,
		Voltage:          laws.CellRestingVoltage,
		activating:       false,
		DendriteSynapses: make(map[SynapseID]bool),
		AxonSynapses:     make(map[SynapseID]bool),
		WasFired:         false,
		OnFired:          make([]func(CellID), 0),
	}
	cell := &c
	network.Cells = append(network.Cells, cell)
	network.cellMux.Unlock()
	return cell
}

/*
FireActionPotential does an action potential cycle.

`cell.activating` property is set in `Step()`
*/
func (cell *Cell) FireActionPotential() {
	cell.WasFired = true
	// log.Println("Action Potential Firing\n  cell=", cell.ID, "synapses=", len(cell.AxonSynapses))

	for _, cb := range cell.OnFired {
		cb(cell.ID)
	}

	for synapseID := range cell.AxonSynapses {
		cell.Network.AddSynapseToNextStep(synapseID)
	}
}

func (cell *Cell) String() string {
	s := fmt.Sprintf("Cell %d", cell.ID)
	s += fmt.Sprintf("\n  Immortal=%t", cell.Immortal)
	s += fmt.Sprintf("\n  Voltage=%d", cell.Voltage)
	s += fmt.Sprintf("\n  Tag=%s", cell.Tag)

	s += fmt.Sprintf("\n  AxonSynapses (%d)", len(cell.AxonSynapses))
	for id := range cell.AxonSynapses {
		s += fmt.Sprintf("\n    %d", id)
	}
	s += fmt.Sprintf("\n  DendriteSynapses (%d)", len(cell.DendriteSynapses))
	for id := range cell.DendriteSynapses {
		s += fmt.Sprintf("\n    %d", id)
	}

	return s
}

func (cell *Cell) addDendrite(synapseID SynapseID) {
	synapse := cell.Network.GetSyn(synapseID)
	synapse.ToNeuronDendrite = cell.ID
	cell.DendriteSynapses[synapseID] = true
}

func (cell *Cell) addAxon(synapseID SynapseID) {
	synapse := cell.Network.GetSyn(synapseID)
	synapse.FromNeuronAxon = cell.ID
	cell.AxonSynapses[synapseID] = true
}

/*
PruneCell removes a cell and its synapses. It is independent of PruneSynapse.
*/
func (network *Network) PruneCell(cellID CellID) {
	// log.Println("Pruning cell", cellID)
	cell := network.GetCell(cellID)
	// with good code, the following should not be necessary.
	if len(cell.DendriteSynapses) > 0 {
		panic(fmt.Sprintf("Should not need to prune dendrite synapse from cell=%d (%v)", cellID, cell.DendriteSynapses))
	}
	if len(cell.AxonSynapses) > 0 {
		panic(fmt.Sprintf("Should not need to prune axon synapse from cell=%d (%v)", cellID, cell.AxonSynapses))
	}

	// Do this after removing synapses, because otherwise we can end up with
	// orphan synapses.
	if cell.Immortal {
		return
	}

	network.cellMux.Lock()
	network.Cells[cellID] = nil
	network.cellMux.Unlock()
}

/*
postRefractoryReset does things that happen after a cell has been through
the refractory period after it fired.
*/
func (cell *Cell) postRefractoryReset() {
	cell.activating = false
	cell.Voltage = laws.CellRestingVoltage
}
