package potential

import (
	"bleh/potential"
	"fmt"
	"strings"
	"time"
)

/*
Sample activates the network with the seed data and replies with
the network's answer.

`start` and `end` will indicate the beginning and end of when to sample data.
*/
func Sample(seed string, data Dataset, network *Network, maxInterations int, start interface{}, end interface{}) string {
	network.Disabled = false
	network.ResetForTraining()
	seedChars := strings.Split(seed, "")
	out := ""
	outConduit := make(chan string)

	// add a callback to each output cell that sends the
	go func() {
		for _, v := range data.KeyToItem {
			network.Cells[v.OutputCell].OnFired = append(
				network.Cells[v.OutputCell].OnFired,
				func(cell potential.CellID) {
					s := data.cellToKey[cell]
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
				out += s
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
	fmt.Println(out)

	// reset all the output cell callbacks
	network.Disabled = true
	for _, v := range data.KeyToItem {
		network.Cells[v.OutputCell].OnFired = make([]func())
	}

	return out
}
