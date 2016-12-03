package potential

type integrityReport struct {
	cellHasMissingAxonSynapse     map[CellID]SynapseID
	cellHasMissingDendriteSynapse map[CellID]SynapseID
	synapseHasMissingDendriteCell map[SynapseID]CellID
	synapseHasMissingAxonCell     map[SynapseID]CellID
}

func newIntegrityReport() integrityReport {
	return integrityReport{
		cellHasMissingAxonSynapse:     make(map[CellID]SynapseID),
		cellHasMissingDendriteSynapse: make(map[CellID]SynapseID),
		synapseHasMissingDendriteCell: make(map[SynapseID]CellID),
		synapseHasMissingAxonCell:     make(map[SynapseID]CellID),
	}
}

func (report *integrityReport) isOK() bool {
	return len(report.cellHasMissingAxonSynapse) == 0 && len(report.cellHasMissingDendriteSynapse) == 0 && len(report.synapseHasMissingAxonCell) == 0 && len(report.synapseHasMissingDendriteCell) == 0
}

func checkIntegrity(network *Network) (bool, integrityReport) {
	report := newIntegrityReport()

	for cellID, cell := range network.Cells {
		for synapseID := range cell.AxonSynapses {
			_, ok := network.Synapses[synapseID]
			if !ok {
				report.cellHasMissingAxonSynapse[cellID] = synapseID
			}
		}
		for synapseID := range cell.DendriteSynapses {
			_, ok := network.Synapses[synapseID]
			if !ok {
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
