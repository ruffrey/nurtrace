package potential

import "time"

// From standard neuroscience Membrane Potential, which conveniently fits in a tiny 8 bit integer.
const apResting int8 = -70
const apThreshold int16 = -55 // needed for comparisons
const apPeak int8 = 40
const apLow int8 = -90

// How much a synapse's ActivationHistory should be incremented extra when
// its firing results in activation (strengthening the synapse)
const synapseAPBoost uint = 1

/*
Cell holds voltage, receives input from Dendrites, and upon reaching the activation voltage,
fires an action potential cycles and its axon synapses push voltage to the dendrites it connects
to.
*/
type Cell struct {
	Voltage               int8
	Activating            bool
	DendriteSynapses      []*Synapse // This cell's inputs.
	AxonSynapses          []*Synapse // this cell's outputs.
	equilibriumTicker     *time.Ticker
	equilibriumQuitSignal chan struct{}
}

/*
NewCell instantiates a Cell
*/
func NewCell() Cell {
	cell := Cell{
		Voltage: apResting,
	}
	cell.startEquilibrium()
	return cell
}

/*
Destroy removes things that might otherwise leak.
*/
func (cell *Cell) Destroy() {
	cell.stopEquilibrium()
}

/*
StartEquilibrium starts the equilibrium cycle, which moves the cell towards equilibrium at
the normal resting potential `apResting`.
*/
func (cell *Cell) startEquilibrium() {
	cell.equilibriumTicker = time.NewTicker(50)
	cell.equilibriumQuitSignal = make(chan struct{})
	go func() {
		for {
			select {
			case <-cell.equilibriumTicker.C:
				if cell.Activating {
					return
				}
				// move it halfway towards its resting potential, every so often. this was
				// arbitrarily chosen.
				halfDiff := (cell.Voltage - apResting) % 2
				cell.Voltage -= halfDiff
				return
			case <-cell.equilibriumQuitSignal:
				cell.equilibriumTicker.Stop()
				return
			}
		}
	}()
}

func (cell *Cell) stopEquilibrium() {
	close(cell.equilibriumQuitSignal)
}

/*
FireActionPotential does an action potential cycle.
*/
func (cell *Cell) FireActionPotential() {
	cell.Activating = true
	time.AfterFunc(1*time.Millisecond, func() {
		cell.Voltage = apPeak // probably not doing anything...hmm.
		for _, synapse := range cell.AxonSynapses {
			synapse.Activate()
		}
		time.AfterFunc(1*time.Millisecond, func() {
			cell.Voltage = apLow
			time.AfterFunc(1*time.Millisecond, func() {
				// other neuron firings may have already bumped this to, or above, the
				// resting potential.
				if cell.Voltage < apResting {
					cell.Voltage = apResting
				}
			})
		})
	})
}

/*
applyVoltage changes it this much, but keeps it between the
*/
func (cell *Cell) applyVoltage(change int8, fromSynapse *Synapse) {
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
