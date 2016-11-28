package potential

import (
	"math/rand"
	"time"
)

/*
apResting comes from standard neuroscience Membrane Potential. This, and all voltages
in the lib, conveniently fit in tiny 8 bit integers.
*/
const apResting int8 = -70

/*
apThreshold represents the millivolts where an action potential will result.
int16 is needed for comparisons.
*/
const apThreshold int16 = -55

/*
synapseAPBoost is how much a synapse's ActivationHistory should be incremented extra when
its firing results in activation (strengthening the synapse)
*/
const synapseAPBoost uint = 1

/*
synapseEffectDelayMillis is the time between another cell's axon firing and the
cell at the end of the synapse getting a voltage boost. The primary reason for
this delay is to normalize timing across all machines. Without it, faster
machines will process voltage changes faster, and a network trained on one
set of hardware will not be usable on another set.
*/
const synapseEffectDelayMillis = 1

/*
refractoryPeriodMillis represents after a neuron fires, the amount of time (ms) is will
be blocked from firing again.
*/
const refractoryPeriodMillis = 4

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
	Network    *Network `json:"-"` // skip circular reference in JSON
	Voltage    int8     // unnecessary to recreate cell
	Activating bool     // unnecessary to recreate cell
	/*
	  DendriteSynapses are this cell's inputs. They are IDs of synapses.
	*/
	DendriteSynapses map[SynapseID]bool
	/*
	  DendriteSynapses are this cell's outputs. They are IDs of synapses. When it fires,
	  these synapses will be triggered.
	*/
	AxonSynapses map[SynapseID]bool
}

/*
NewCell instantiates a Cell *and* adds it to the network's list of cells.
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
	network.Cells[cell.ID] = &cell
	return cell
}

/*
FireActionPotential does an action potential cycle.
*/
func (cell *Cell) FireActionPotential() {
	// fmt.Println("Action Potential Firing\n  cell=", cell.ID, "\n  axon synapses=", cell.AxonSynapses)
	cell.Activating = true

	// activate all synapses on its axon
	for synapseID := range cell.AxonSynapses {
		synapse := cell.Network.Synapses[synapseID]
		// fmt.Println("  activating synapse", synapse, "from cell", cell.ID)
		synapse.Activate()
	}

	time.AfterFunc(refractoryPeriodMillis*time.Millisecond, func() {
		cell.Voltage = apResting
		cell.Activating = false
	})
}

/*
ApplyVoltage changes the cell's voltage by a specified amount much.
Care is taken to prevent the tiny int8 variables from overflowing.
Voltage may not change for a few milliseconds depending on `synapseEffectDelayMillis`.
*/
func (cell *Cell) ApplyVoltage(change int8, fromSynapse *Synapse) {
	time.AfterFunc(synapseEffectDelayMillis*time.Millisecond, func() {
		if cell.Activating {
			// Block during action potential cycle
			return
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
		}
		cell.Voltage = int8(newPossibleVoltage)
	})
}
