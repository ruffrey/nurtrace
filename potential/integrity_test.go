package potential

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Integrity(t *testing.T) {
	t.Run("isOK() works", func(t *testing.T) {
		report := newIntegrityReport()
		assert.Equal(t, true, report.isOK())
	})
	t.Run("report.Print works", func(t *testing.T) {
		report := newIntegrityReport()
		report.Print()
	})
	t.Run("cell with bad dentrite synapse", func(t *testing.T) {
		network := NewNetwork()
		cell := NewCell(network)
		cell.DendriteSynapses[999] = true
		ok, report := CheckIntegrity(network)
		assert.Equal(t, false, ok)
		assert.Equal(t, 1, len(report.cellHasMissingDendriteSynapse))
		assert.Equal(t, 0, len(report.synapseHasMissingDendriteCell))
		assert.Equal(t, 0, len(report.synapseHasMissingAxonCell))
		assert.Equal(t, 0, len(report.cellHasMissingAxonSynapse))
		assert.Equal(t, SynapseID(999), report.cellHasMissingDendriteSynapse[cell.ID])
	})
	t.Run("cell with bad axon synapse", func(t *testing.T) {
		network := NewNetwork()
		cell := NewCell(network)
		cell.AxonSynapses[7777] = true
		ok, report := CheckIntegrity(network)
		assert.Equal(t, false, ok)
		assert.Equal(t, 1, len(report.cellHasMissingAxonSynapse))
		assert.Equal(t, 0, len(report.synapseHasMissingDendriteCell))
		assert.Equal(t, 0, len(report.synapseHasMissingAxonCell))
		assert.Equal(t, 0, len(report.cellHasMissingDendriteSynapse))
		assert.Equal(t, SynapseID(7777), report.cellHasMissingAxonSynapse[cell.ID])
	})
	t.Run("synapse with bad dendrite cell", func(t *testing.T) {
		network := NewNetwork()
		synapse := NewSynapse(network)
		cell := NewCell(network)
		synapse.FromNeuronAxon = cell.ID
		synapse.ToNeuronDendrite = 1
		ok, report := CheckIntegrity(network)
		assert.Equal(t, false, ok)
		assert.Equal(t, 1, len(report.synapseHasMissingDendriteCell))
		assert.Equal(t, 0, len(report.synapseHasMissingAxonCell))
		assert.Equal(t, 0, len(report.cellHasMissingAxonSynapse))
		assert.Equal(t, 0, len(report.cellHasMissingDendriteSynapse))
		assert.Equal(t, CellID(1), report.synapseHasMissingDendriteCell[synapse.ID])
	})
	t.Run("synapse with bad axon cell", func(t *testing.T) {
		network := NewNetwork()
		synapse := NewSynapse(network)
		cell := NewCell(network)
		synapse.FromNeuronAxon = 2
		synapse.ToNeuronDendrite = cell.ID
		ok, report := CheckIntegrity(network)
		assert.Equal(t, false, ok)
		assert.Equal(t, 0, len(report.synapseHasMissingDendriteCell))
		assert.Equal(t, 1, len(report.synapseHasMissingAxonCell))
		assert.Equal(t, 0, len(report.cellHasMissingAxonSynapse))
		assert.Equal(t, 0, len(report.cellHasMissingDendriteSynapse))
		assert.Equal(t, CellID(2), report.synapseHasMissingAxonCell[synapse.ID])
	})
}
