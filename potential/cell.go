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
	return CellID(rand.Int())
}

/*
Cell holds voltage, receives input from Dendrites, and upon reaching the activation voltage,
fires an action potential cycles and its axon synapses push voltage to the dendrites it connects
to.
*/
type Cell struct {
	ID         CellID
	Network    *Network
	Voltage    int8
	Activating bool
	/*
		DendriteSynapses are this cell's inputs. They are IDs of synapses.
	*/
	DendriteSynapses  []SynapseID
	AxonSynapses      []SynapseID // this cell's outputs.
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
		DendriteSynapses: make([]SynapseID, 0),
		AxonSynapses:     make([]SynapseID, 0),
	}
	return cell
}

/*
FireActionPotential does an action potential cycle.
*/
func (cell *Cell) FireActionPotential() {
	// fmt.Println("Action Potential firing", cell.ID, "for synapses", cell.AxonSynapses)
	cell.Activating = true
	cell.Voltage = apPeak // probably not doing anything...hmm.

	// activate all synapses on its axon
	for _, synapseID := range cell.AxonSynapses {
		synapse := cell.Network.Synapses[synapseID]
		synapse.Activate()
	}
	time.AfterFunc(10*time.Millisecond, func() {
		cell.Voltage = apLow
		time.AfterFunc(10*time.Millisecond, func() {
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
