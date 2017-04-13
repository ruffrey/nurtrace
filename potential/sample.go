package potential

import (
	"time"
)

/*
Sample activates the network with the seed data and replies with
the network's answer.

`start` and `end` will indicate the beginning and end of when to sample data.
*/
func Sample(seedKeys []interface{}, data *Dataset, network *Network, maxResultLen int, start interface{}, end interface{}) []interface{} {
	network.Disabled = false
	network.ResetForTraining()
	var out []interface{}
	outConduit := make(chan interface{})
	maxTimeout := time.Second * 1 // do not let samples run longer than this

	// add a callback to each output cell that sends the result back
	for _, v := range data.KeyToItem {
		if v.OutputCell == 0 {
			continue
		}
		network.Cells[v.OutputCell].OnFired = append(
			network.Cells[v.OutputCell].OnFired,
			func(cell CellID) {
				s := data.CellToKey[cell]
				if s == start {
					return
				}
				outConduit <- s
			},
		)
	}

	// wait for the output to come through the network
	done := make(chan bool)
	go func() {
		for {
			select {
			case s := <-outConduit:
				if s == end {
					done <- true
					break
				}
				// fmt.Print(s)
				out = append(out, s)
				if len(out) >= maxResultLen {
					done <- true
					break
				}
			case <-time.After(maxTimeout):
				done <- true
				break
			}
		}

	}()

	// fire one cell per step, in order.
	if start != nil {
		ic := data.KeyToItem[start].InputCell
		network.Cells[ic].FireActionPotential()
		network.resetCellsOnNextStep[ic] = true
		network.Step()
	}

	for _, perceptionUnit := range seedKeys {
		ic := data.KeyToItem[perceptionUnit].InputCell
		network.Cells[ic].FireActionPotential()
		network.resetCellsOnNextStep[ic] = true
	}

	go func() {
		// convert to bool for use in randCellFromMap
		randCellList := make(map[CellID]bool)
		for cellID := range data.CellToKey {
			randCellList[cellID] = true
		}
		for i := 0; i < 10000; i++ {
			r := randCellFromMap(randCellList)
			network.getCell(r).FireActionPotential()
			network.resetCellsOnNextStep[r] = true
		}
		for {
			hasMore := network.Step()
			if !hasMore {
				return
			}
		}

	}()

	<-done

	// reset all the output cell callbacks
	network.Disabled = true
	for _, v := range data.KeyToItem {
		if v.OutputCell != 0 {
			network.Cells[v.OutputCell].OnFired = make([]func(CellID), 0)
		}
	}

	return out
}
