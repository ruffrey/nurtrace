package potential

import "time"

/*
Sample activates the network with the seed data and replies with
the network's answer.

`start` and `end` will indicate the beginning and end of when to sample data.
*/
func Sample(seedKeys []interface{}, data *Dataset, network *Network, maxInterations int, start interface{}, end interface{}) []interface{} {
	network.Disabled = false
	network.ResetForTraining()
	var out []interface{}
	outConduit := make(chan interface{})
	maxTimeout := time.Second * 1 // do not let samples run longer than this

	// add a callback to each output cell that sends the result back
	go func() {
		for _, v := range data.KeyToItem {
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
	}()

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
				if len(out) >= maxInterations {
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
	network.Cells[data.KeyToItem[start].InputCell].FireActionPotential()
	network.Step()
	for _, perceptionUnit := range seedKeys {
		network.Cells[data.KeyToItem[perceptionUnit].InputCell].FireActionPotential()
		network.Step()
	}

	<-done

	// reset all the output cell callbacks
	network.Disabled = true
	for _, v := range data.KeyToItem {
		network.Cells[v.OutputCell].OnFired = make([]func(CellID), 0)
	}

	return out
}
