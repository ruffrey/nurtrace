package potential

import (
	"fmt"
	"math/rand"
)

/*
CellID should be unique for all cells in a network.
*/
type CellID uint32

/*
NewCellID makes a new random CellID.
*/
func NewCellID() (cid CellID) {
	return CellID(rand.Uint32())
}

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
	Voltage    int8     // unnecessary to recreate cell
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
	  WasFired is used during training to know if this cell fired during the session
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
	var id CellID
	for {
		id = NewCellID()
		if _, alreadyExists := network.Cells[id]; !alreadyExists {
			break
		}
		// fmt.Println("warn: would have gotten dupe cell ID")
	}
	c := Cell{
		ID:               id,
		Network:          network,
		Immortal:         false,
		Voltage:          apResting,
		activating:       false,
		DendriteSynapses: make(map[SynapseID]bool),
		AxonSynapses:     make(map[SynapseID]bool),
		WasFired:         false,
		OnFired:          make([]func(CellID), 0),
	}
	cell := &c
	network.cellMux.Lock()
	network.Cells[cell.ID] = cell
	network.cellMux.Unlock()
	return cell
}

/*
FireActionPotential does an action potential cycle.
*/
func (cell *Cell) FireActionPotential() {
	cell.WasFired = true
	cell.activating = true

	if cell.Network.Disabled {
		return
	}
	// fmt.Println("Action Potential Firing\n  cell=", cell.ID, "syanpses=", len(cell.AxonSynapses))

	for _, cb := range cell.OnFired {
		go cb(cell.ID)
	}

	for synapseID := range cell.AxonSynapses {
		if cell.Network.Disabled {
			// This will likely happen when setting the network to disabled, because
			// timeouts will still need to wait to finish then try to fire the network.
			// While technically it should be ok to let the network fizzle out on its
			// own in the ApplyVoltage check for network being disabled, stopping here
			// will save some precious CPU.
			// fmt.Println("warn: network stopped during FireActionPotential")
			break
		}
		cell.Network.AddSynapseToNextStep(synapseID)
	}
}

/*
ApplyVoltage changes the cell's voltage by a specified amount much.

Care is taken to prevent the tiny int8 variables from overflowing.
*/
func (cell *Cell) ApplyVoltage(change int8, fromSynapse *Synapse) (didFire bool) {
	didFire = false

	// Block during action potential cycle or when network is disabled.
	if cell.activating || cell.Network.Disabled {
		// disable any more voltage applications from cells once the network has been disabled,
		// which will let the network firings sizzle out after a refractory period.
		return didFire
	}

	// Neither alone will be outside int8 bounds, but we need to prevent
	// possible int8 buffer overflow in the result.
	var newPossibleVoltage int16
	newPossibleVoltage = int16(change) + int16(cell.Voltage)
	if newPossibleVoltage > apThreshold {
		// when a synapse firing results in firing an Action Potential, it counts toward making
		// the synapse stronger, so we increment the ActivationHistory a second time
		fromSynapse.ActivationHistory += synapseAPBoost
		cell.FireActionPotential()
		didFire = true
	}
	cell.Voltage = int8(newPossibleVoltage)

	return didFire
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
