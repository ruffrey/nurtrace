package potential

import (
	"strings"
	"time"
)

/*
Sample activates the network with the seed data and replies with
the network's answer.

`start` and `end` will indicate the beginning and end of when to sample data.
*/
func Sample(seed string, data *Dataset, network *Network, maxInterations int, start interface{}, end interface{}) []interface{} {
	network.Disabled = false
	network.ResetForTraining()
	seedChars := strings.Split(seed, "")
	var out []interface{}
	outConduit := make(chan interface{})
	maxTimeout := time.Second * 5 // do not let samples run longer than this

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

	network.Cells[data.KeyToItem[start].InputCell].FireActionPotential()
	network.Step()
	for _, char := range seedChars {
		network.Cells[data.KeyToItem[char].InputCell].FireActionPotential()
		network.Step()
	}

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
				hasMore := network.Step()
				if !hasMore {
					done <- true
					break
				}
			case <-time.After(maxTimeout):
				done <- true
			}
		}

	}()
	<-done

	// reset all the output cell callbacks
	network.Disabled = true
	for _, v := range data.KeyToItem {
		network.Cells[v.OutputCell].OnFired = make([]func(CellID), 0)
	}

	return out
}
