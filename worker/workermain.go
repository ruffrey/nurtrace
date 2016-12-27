package main

import (
	"bleh/potential"
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
