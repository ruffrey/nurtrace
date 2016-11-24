package potential

import (
	"math/rand"
	"time"
)

// From standard neuroscience Membrane Potential, which conveniently fits in a tiny 8 bit integer.
const apResting int8 = -70
const apThreshold int16 = -55 // needed for comparisons
const apPeak int8 = 40
const apLow int8 = -90

// How much a synapse's ActivationHistory should be incremented extra when
// its firing results in activation (strengthening the synapse)
const synapseAPBoost uint = 1

/*
CellID is a normal Go integer that should be unique for all cells in a network.
*/
type CellID int

/*
NewCellID makes a new random CellID.
*/
func NewCellID() (cid CellID) {
	i := rand.Int()
	if i == 0 {
		panic("Should never get zero from rand.Int()")
	}
	return CellID(i)
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
	ID         CellID
	Network    *Network
	Voltage    int8
	Activating bool
	/*
	  DendriteSynapses are this cell's inputs. They are IDs of synapses.
	*/
	DendriteSynapses map[SynapseID]bool
	/*
	  DendriteSynapses are this cell's outputs. They are IDs of synapses. When it fires,
	  these synapses will be triggered.
	*/
	AxonSynapses      map[SynapseID]bool
	equilibriumTicker *time.Timer
}

/*
NewCell instantiates a Cell
*/
func NewCell(network *Network) Cell {
	cell := Cell{
		ID:               NewCellID(),
		Network:          network,
		Voltage:          apResting,
		Activating:       false,
		DendriteSynapses: make(map[SynapseID]bool),
		AxonSynapses:     make(map[SynapseID]bool),
	}
	return cell
}

/*
FireActionPotential does an action potential cycle.
*/
func (cell *Cell) FireActionPotential() {
	// fmt.Println("Action Potential Firing\n  cell=", cell.ID, "\n  axon synapses=", cell.AxonSynapses)
	cell.Activating = true
	cell.Voltage = apPeak // probably not doing anything...hmm.

	// activate all synapses on its axon
	for synapseID := range cell.AxonSynapses {
		synapse := cell.Network.Synapses[synapseID]
		// fmt.Println("  activating synapse", synapse, "from cell", cell.ID)
		synapse.Activate()
	}
	time.AfterFunc(4*time.Millisecond, func() {
		cell.Voltage = apLow
		time.AfterFunc(4*time.Millisecond, func() {
			// other neuron firings may have already bumped this to, or above, the
			// resting potential.
			if cell.Voltage < apResting {
				cell.Voltage = apResting
			}
		})
	})
}

/*
ApplyVoltage changes it this much, but keeps it between the
*/
func (cell *Cell) ApplyVoltage(change int8, fromSynapse *Synapse) {
	if cell.Activating {
		// Block during action potential cycle
		return
	}

	// Neither alone will be outside int8 bounds, but we need to prevent
	// possible int8 buffer overflow in the result.
	newPossibleVoltage := int16(change) + int16(cell.Voltage)
	if newPossibleVoltage > apThreshold {
		// when a synapse firing results in firing an Action Potential, it counts toward making
		// the synapse stronger, so we increment the ActivationHistory a second time
		fromSynapse.ActivationHistory += synapseAPBoost
		cell.FireActionPotential()
	}
	cell.Voltage = int8(newPossibleVoltage)
}
