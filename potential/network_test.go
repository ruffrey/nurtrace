package potential

import (
	"github.com/ruffrey/nurtrace/laws"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewNetwork(t *testing.T) {
	original := NewNetwork()

	t.Run("Cell map is initialized and can add cell immediately", func(t *testing.T) {
		cell := NewCell(original)
		original.Cells[cell.ID] = cell // already done in factory
	})

	t.Run("Synapse map is initialized and can add synapse immediately", func(t *testing.T) {
		synapse := NewSynapse(original)
		original.Synapses[synapse.ID] = synapse // already done in factory
	})

	t.Run("Calling Grow() with all empty values does not crash", func(t *testing.T) {
		original.Grow(0, 0, 0)
	})

	t.Run("new network passes integrity check", func(t *testing.T) {
		network := NewNetwork()
		network.Grow(100, 10, 100)
		ok, _ := CheckIntegrity(network)
		assert.Equal(t, true, ok, "new network has integrity issues")
	})
}

func Test_BasicNetworkFiring(t *testing.T) {
	t.Run("does not panic during forced FireActionPotential", func(t *testing.T) {
		network := NewNetwork()
		neuronsToAdd := 50
		synapsesToAdd := 100
		network.Grow(neuronsToAdd, 10, synapsesToAdd)
		iterations := 1000
		for i := 1; i < iterations; i++ {
			cellID := network.RandomCellKey()
			cell := network.Cells[cellID]
			cell.FireActionPotential()
			network.Step()
		}
		network.Step()
	})
}

func Test_NetworkSerialization(t *testing.T) {
	var network *Network
	var synapse *Synapse
	var cell1, cell2 *Cell
	before := func() {
		// setup
		network = NewNetwork()
		synapse = NewSynapse(network)
		// cell 1 fires into cell 2
		cell1 = NewCell(network)
		cell2 = NewCell(network)
		cell1.AxonSynapses[synapse.ID] = true
		cell2.DendriteSynapses[synapse.ID] = true
		synapse.FromNeuronAxon = cell1.ID
		synapse.ToNeuronDendrite = cell2.ID
	}

	t.Run("serializes network to JSON despite having circular references", func(t *testing.T) {
		before()
		json, err := network.ToJSON()
		assert.Equal(t, err, nil, "error occurred generating json from network")
		assert.NotEqual(t, json, "", "json failed to be generated from network")
	})
	t.Run("writes serialized network to disk, reads back and hydrates all pointers", func(t *testing.T) {
		before()
		err := network.SaveToFile("_network.test.json")
		if err != nil {
			assert.NoError(t, err, "Error testing network serialization to file")
		}
		net2, err := LoadNetworkFromFile("_network.test.json")
		if err != nil {
			assert.NoError(t, err, "Failed reading network from save file")
		}
		n1, err := network.ToJSON()
		assert.NoError(t, err)
		n2, err := net2.ToJSON()
		assert.NoError(t, err)
		assert.EqualValues(t, n1, n2, "loaded network does not match original")

		ok, report := CheckIntegrity(net2)
		assert.Equal(t, true, ok, report)
	})

	t.Run("loading a network that has bad integrity returns an error", func(t *testing.T) {
		// setup
		filepath := "_bad_net_integ.test.json"
		n := NewNetwork()
		c := NewCell(n)
		c.AxonSynapses[5] = true
		s := NewSynapse(n)
		s.FromNeuronAxon = c.ID
		s.ToNeuronDendrite = 11
		assert.Equal(t, 1, len(n.Cells))
		assert.Equal(t, 1, len(n.Synapses))
		integrity, _ := CheckIntegrity(n)
		assert.Equal(t, false, integrity,
			"pre-setup network should purposely have bad integrity")
		err := n.SaveToFile(filepath)
		assert.NoError(t, err, "pre-setup network failed")
		// test
		_, err = LoadNetworkFromFile(filepath)
		assert.Error(t, err, "loading a network with poor integrity did not return error")
		if err == nil {
			t.Fail()
			return
		}
		assert.Equal(t, err.Error(),
			"Cannot load network with bad integrity from file "+filepath)
	})

	t.Run("loading a non-existant file network returns an error", func(t *testing.T) {
		_, err := LoadNetworkFromFile("/kasdjfkk/asdkfjdsk")
		assert.Error(t, err)
		assert.NotEmpty(t, err)
	})

	t.Run("loading a network with bad json returns an error", func(t *testing.T) {
		filepath := "_net_test_notjson.test.json"
		err := ioutil.WriteFile(filepath, []byte("asdf not json"), os.ModePerm)
		assert.NoError(t, err)
		_, err = LoadNetworkFromFile(filepath)
		assert.Error(t, err)
		assert.NotEmpty(t, err)
	})

}

func Test_LinkCells(t *testing.T) {
	t.Run("linkCells adds synapse fromCell fires toCell", func(t *testing.T) {
		network := NewNetwork()
		fromCell := NewCell(network)
		toCell := NewCell(network)

		synapse := network.linkCells(fromCell.ID, toCell.ID)

		assert.Equal(t, synapse.FromNeuronAxon, fromCell.ID)
		assert.Equal(t, synapse.ToNeuronDendrite, toCell.ID)
		assert.Equal(t, true, fromCell.AxonSynapses[synapse.ID])
		assert.Equal(t, true, toCell.DendriteSynapses[synapse.ID])
	})
}

func Test_ResetForTraining(t *testing.T) {
	t.Run("resets cell props activating, wasFired, but NOT voltage", func(t *testing.T) {
		network := NewNetwork()

		// pretest
		cell1 := NewCell(network)
		assert.Equal(t, false, cell1.activating)
		assert.Equal(t, int16(laws.CellRestingVoltage), cell1.Voltage)
		assert.Equal(t, false, cell1.WasFired)

		// setup
		cell1.activating = true
		cell1.Voltage = int16(100)
		cell1.WasFired = true

		network.Cells[cell1.ID] = cell1

		network.ResetForTraining()

		// assertions
		assert.Equal(t, false, cell1.activating)
		assert.Equal(t, int16(100), cell1.Voltage)
		assert.Equal(t, false, cell1.WasFired)
	})

	// To make sure we don't do this accidentally because it seems like the right thing,
	// and break synapse pruning.
	t.Run("does not reset synapse ActivationHistory", func(t *testing.T) {
		network := NewNetwork()

		s := NewSynapse(network)
		s.ActivationHistory = 12

		network.ResetForTraining()

		assert.Equal(t, uint(12), network.Synapses[s.ID].ActivationHistory)
	})
}

func Test_NetworkPrint(t *testing.T) {
	t.Run("network.Print works", func(t *testing.T) {
		network := NewNetwork()
		network.Grow(5, 2, 5)
		network.Print()
	})
	t.Run("network.PrintTotals works", func(t *testing.T) {
		network := NewNetwork()
		network.Grow(5, 2, 5)
		network.PrintTotals()
	})
}
