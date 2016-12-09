package potential

import (
	"fmt"
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

	// add a callback to each output cell that sends the
	go func() {
		for _, v := range data.KeyToItem {
			network.Cells[v.OutputCell].OnFired = append(
				network.Cells[v.OutputCell].OnFired,
				func(cell CellID) {
					s := data.CellToKey[cell]
					fmt.Println(cell, s)
					if s == start {
						return
					}
					outConduit <- s
				},
			)
		}
	}()

	go func() {
		network.Cells[data.KeyToItem[start].InputCell].FireActionPotential()
		time.AfterFunc(sleepBetweenInputTriggers, func() {
			for _, char := range seedChars {
				ch := make(chan bool)
				time.AfterFunc(sleepBetweenInputTriggers, func() {
					network.Cells[data.KeyToItem[char].InputCell].FireActionPotential()
					ch <- true
				})
				<-ch
			}
		})
	}()

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
			case <-time.After(time.Second * 3):
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
