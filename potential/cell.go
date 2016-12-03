package potential

import (
	"fmt"
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
SynapseEffectDelayMicrosecs is the time between another cell's axon firing and the
cell at the end of the synapse getting a voltage boost. The primary reason for
this delay is to normalize timing across all machines. Without it, faster
machines will process voltage changes faster, and a network trained on one
set of hardware will not be usable on another set.
*/
const SynapseEffectDelayMicrosecs = 100
const synapseDelay = SynapseEffectDelayMicrosecs * time.Microsecond

/*
RefractoryPeriodMillis represents after a neuron fires, the amount of time (ms) is will
be blocked from firing again.
*/
const RefractoryPeriodMillis = 4
const refractory = RefractoryPeriodMillis * time.Millisecond

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

It is up to the implementer to set the Network pointer and add it to the network.
*/
func NewCell() *Cell {
	cell := Cell{
		ID:               NewCellID(),
		Immortal:         false,
		Voltage:          apResting,
		activating:       false,
		DendriteSynapses: make(map[SynapseID]bool),
		AxonSynapses:     make(map[SynapseID]bool),
		WasFired:         false,
		OnFired:          make([]func(CellID), 0),
	}
	return &cell
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

	done := make(chan bool)
	// activate all synapses on its axon
	// using a goroutine seems to help avoid crashes where there are concurrent
	// reads and writes of the same map. That would happen if
	go func() {
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
			synapse, exists := cell.Network.Synapses[synapseID]
			if !exists {
				fmt.Println("error: synapse", synapseID, "does not exist on cell", cell.ID)
				continue
			}
			// fmt.Println("  activating synapse", synapse, "\n  from cell", cell.ID)
			err := synapse.Activate()
			if err != nil {
				fmt.Println("cell fire err:", err)
			}
		}
		done <- true
	}()

	time.AfterFunc(refractory, func() {
		cell.Voltage = apResting
		cell.activating = false
	})
	<-done
}

/*
ApplyVoltage changes the cell's voltage by a specified amount much.
Care is taken to prevent the tiny int8 variables from overflowing.
Voltage may not change for a few milliseconds depending on `SynapseEffectDelayMicrosecs`.
*/
func (cell *Cell) ApplyVoltage(change int8, fromSynapse *Synapse) {
	// fmt.Println("ApplyVoltage", cell.ID, cell.Network.Disabled, &cell.Network)
	if cell.Network.Disabled {
		// disable more voltage applications from cells once the network has been disabled,
		// which will let the network firings sizzle out after a refractory period or so.
		// fmt.Println("warn: attempt to apply voltage when network disabled")
		return
	}

	time.AfterFunc(synapseDelay, func() {
		if cell.activating {
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
