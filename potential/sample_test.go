package potential

import "testing"

func Test_Sampling(t *testing.T) {
	t.Run("sampling works", func(t *testing.T) {
		var seedKeys []interface{}
		seedKeys = append(seedKeys, "START")
		seedKeys = append(seedKeys, "START")
		seedKeys = append(seedKeys, "START")
		seedKeys = append(seedKeys, "START")
		seedKeys = append(seedKeys, "START")
		seedKeys = append(seedKeys, "START")
		seedKeys = append(seedKeys, "START")
		seedKeys = append(seedKeys, "START")
		seedKeys = append(seedKeys, "START")
		seedKeys = append(seedKeys, "START")
		seedKeys = append(seedKeys, "START")
		seedKeys = append(seedKeys, "START")
		seedKeys = append(seedKeys, "START")
		seedKeys = append(seedKeys, "START")
		seedKeys = append(seedKeys, "START")
		seedKeys = append(seedKeys, "START")
		seedKeys = append(seedKeys, "START")
		seedKeys = append(seedKeys, "A")
		seedKeys = append(seedKeys, "A")
		seedKeys = append(seedKeys, "A")
		seedKeys = append(seedKeys, "A")
		seedKeys = append(seedKeys, "A")
		seedKeys = append(seedKeys, "A")
		seedKeys = append(seedKeys, "A")
		seedKeys = append(seedKeys, "A")
		seedKeys = append(seedKeys, "A")
		seedKeys = append(seedKeys, "A")
		seedKeys = append(seedKeys, "A")
		seedKeys = append(seedKeys, "A")
		seedKeys = append(seedKeys, "A")
		seedKeys = append(seedKeys, "A")
		seedKeys = append(seedKeys, "A")
		data := Dataset{
			KeyToItem: make(map[interface{}]PerceptionUnit),
			CellToKey: make(map[CellID]interface{}),
		}
		network := NewNetwork()
		network.Grow(20, 5, 0)
		data.KeyToItem["START"] = PerceptionUnit{
			Value:      "START",
			InputCell:  network.RandomCellKey(),
			OutputCell: network.RandomCellKey(),
		}
		data.KeyToItem["A"] = PerceptionUnit{
			Value:      "A",
			InputCell:  network.RandomCellKey(),
			OutputCell: network.RandomCellKey(),
		}
		data.KeyToItem["END"] = PerceptionUnit{
			Value:      "END",
			InputCell:  network.RandomCellKey(),
			OutputCell: network.RandomCellKey(),
		}
		data.CellToKey[data.KeyToItem["START"].InputCell] = "START"
		data.CellToKey[data.KeyToItem["A"].InputCell] = "START"
		data.CellToKey[data.KeyToItem["END"].InputCell] = "END"

		network.GrowPathBetween(data.KeyToItem["START"].InputCell, data.KeyToItem["A"].InputCell, 20)
		network.GrowPathBetween(data.KeyToItem["A"].InputCell, data.KeyToItem["END"].InputCell, 20)

		t.Run("under max iterations", func(t *testing.T) {
			Sample(seedKeys, &data, network, 1000, data.KeyToItem["START"].Value, data.KeyToItem["END"].Value)
		})
		t.Run("over max iterations", func(t *testing.T) {
			Sample(seedKeys, &data, network, 2, data.KeyToItem["START"].Value, data.KeyToItem["END"].Value)
		})
		t.Run("when hitting end", func(t *testing.T) {
			seedKeys = append(seedKeys, "END")
			seedKeys = append(seedKeys, "END")
			seedKeys = append(seedKeys, "END")
			seedKeys = append(seedKeys, "END")
			Sample(seedKeys, &data, network, 1000, "START", "END")
		})

		// also sample with max iterations
	})
}
