package main

import (
	"bleh/perception"
	"bleh/potential"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
)

// CategoryModel is a model that goes from discrete data to a named category
type CategoryModel struct {
	categories []categoryPerceptionUnit
	perception.Perception
}

type categoryPerceptionUnit struct {
	RGB        []uint8
	Name       string
	Hex        string
	OutputCell potential.CellID
}

func readCategeories(filename string) (data []categoryPerceptionUnit, err error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return data, err
	}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return data, err
	}
	return data, err
}

var categoryDataFile = flag.String("data", "", "Filename location of categories to train from")

func main() {
	flag.Parse()
	if *categoryDataFile == "" {
		flag.PrintDefaults()
		return
	}
	cats, err := readCategeories(*categoryDataFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(cats))
	fmt.Println(cats)

	// 256 * 3 possible inputs
	// len(cats) possible outputs
	// Three inputs at a time that should fire the output

	network := potential.NewNetwork()
	neuronsToAdd := 256*3 + len(cats)
	synapsesPerNeuron := 10
	network.GrowRandomNeurons(neuronsToAdd, synapsesPerNeuron)

}
