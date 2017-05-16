package main

import (
	"fmt"
	"math"
	"time"

	"github.com/ruffrey/nurtrace/potential"
)

func main() {
	z := 1000000
	network := potential.NewNetwork()
	m := make(map[int]*potential.Cell, z)
	sl := make([]*potential.Cell, z)

	for i := 0; i < z; i++ {
		m[i] = potential.NewCell(network)
		sl[i] = potential.NewCell(network)
	}

	t0 := time.Now()
	for key, value := range m {
		if key != int(value.ID) { // never happens
			// fmt.Println("m", key, value)
		}
	}
	d0 := time.Now().Sub(t0)

	t1 := time.Now()
	for key, value := range sl {
		if key != int(value.ID) { // never happens
			// fmt.Println("sl", key, value)
		}
	}
	d1 := time.Now().Sub(t1)

	fmt.Println(
		"Iterations:", z,
		"\nmap:", d0,
		"\nslice:", d1,
		"\ndiff:", math.Max(float64(d0), float64(d1))/math.Min(float64(d0), float64(d1)),
	)
}
