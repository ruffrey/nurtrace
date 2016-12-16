package potential

import "fmt"

/*
IntegrityReport lists all bad connections in a network that was tested.
*/
type IntegrityReport struct {
	cellHasMissingAxonSynapse     map[CellID]SynapseID
	cellHasMissingDendriteSynapse map[CellID]SynapseID
	synapseHasMissingDendriteCell map[SynapseID]CellID
	synapseHasMissingAxonCell     map[SynapseID]CellID
}

func newIntegrityReport() IntegrityReport {
	return IntegrityReport{
		cellHasMissingAxonSynapse:     make(map[CellID]SynapseID),
		cellHasMissingDendriteSynapse: make(map[CellID]SynapseID),
		synapseHasMissingDendriteCell: make(map[SynapseID]CellID),
		synapseHasMissingAxonCell:     make(map[SynapseID]CellID),
	}
}

/*
Print outputs the contents of the report to stdout
*/
func (report *IntegrityReport) Print() {
	fmt.Println("cellHasMissingAxonSynapse", report.cellHasMissingAxonSynapse)
	fmt.Println("cellHasMissingDendriteSynapse", report.cellHasMissingDendriteSynapse)
	fmt.Println("synapseHasMissingDendriteCell", report.synapseHasMissingDendriteCell)
	fmt.Println("synapseHasMissingAxonCell", report.synapseHasMissingAxonCell)
}

func (report *IntegrityReport) isOK() bool {
	return len(report.cellHasMissingAxonSynapse) == 0 && len(report.cellHasMissingDendriteSynapse) == 0 && len(report.synapseHasMissingAxonCell) == 0 && len(report.synapseHasMissingDendriteCell) == 0
}

/*
CheckIntegrity tells you whether a network has bad connections between cells.
*/
func CheckIntegrity(network *Network) (bool, IntegrityReport) {
	report := newIntegrityReport()

	for cellID, cell := range network.Cells {
		for synapseID := range cell.AxonSynapses {
			if _, ok := network.Synapses[synapseID]; !ok {
				report.cellHasMissingAxonSynapse[cellID] = synapseID
			}
		}
		for synapseID := range cell.DendriteSynapses {
			if _, ok := network.Synapses[synapseID]; !ok {
				report.cellHasMissingDendriteSynapse[cellID] = synapseID
			}
		}
	}

	for synapseID, synapse := range network.Synapses {
		if _, ok := network.Cells[synapse.FromNeuronAxon]; !ok {
			report.synapseHasMissingAxonCell[synapseID] = synapse.FromNeuronAxon
		}
		if _, ok := network.Cells[synapse.ToNeuronDendrite]; !ok {
			report.synapseHasMissingDendriteCell[synapseID] = synapse.ToNeuronDendrite
		}
	}

	ok := report.isOK()

	return ok, report
}
