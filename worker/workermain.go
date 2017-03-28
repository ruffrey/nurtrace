package main

import (
	"github.com/ruffrey/nurtrace/potential"
	"fmt"
	"os"
)

func main() {
	err := potential.RunWorker()
	if err != nil {
		panic(err)
	}
	hn, _ := os.Hostname()
	fmt.Println("Remote training finished", hn)
}
