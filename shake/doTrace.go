package main

import (
	"os"
	"runtime/trace"
)

func doTrace() {
	f, err := os.Create("trace.out")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err = trace.Start(f); err != nil {
		panic(err)
	}
	defer trace.Stop()
}
