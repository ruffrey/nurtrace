package main

import (
	"fmt"
	"os"

	"github.com/ruffrey/nurtrace/potential"
)

func main() {
	err := potential.RunWorker()
	if err != nil {
		panic(err)
	}
	hn, _ := os.Hostname()
	fmt.Println("Remote training finished", hn)
}
