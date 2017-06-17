package main

import (
	"log"
	"os"

	"github.com/ruffrey/nurtrace/potential"
)

func main() {
	err := potential.RunWorker()
	if err != nil {
		panic(err)
	}
	hn, _ := os.Hostname()
	log.Println("Remote training finished", hn)
}
