package main

import (
	"github.com/ruffrey/nurtrace/potential"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/awalterschulze/gographviz"
)

var networkJSONFile = flag.String("network", "network.json", "Input location of the network JSON")
var outFile = flag.String("out", "network.dot", "Output location of the DOT file")

const directed bool = false

func main() {
	flag.Parse()
	fmt.Println("Visualizing", *networkJSONFile)
	network, err := potential.LoadNetworkFromFile(*networkJSONFile)
	if err != nil {
		panic(err)
	}

	graph := gographviz.NewGraph()

	fmt.Println("- Adding cells", len(network.Cells))
	for _, cell := range network.Cells {
		graph.AddNode("G", strconv.Itoa(int(cell.ID)), nil)
	}
	fmt.Println("- Adding synapses", len(network.Synapses))
	for _, synapse := range network.Synapses {
		graph.AddEdge(strconv.Itoa(int(synapse.FromNeuronAxon)), strconv.Itoa(int(synapse.ToNeuronDendrite)), directed, nil)
	}

	output := graph.String()

	fmt.Println("- Writing to", *outFile)
	err = ioutil.WriteFile(*outFile, []byte(output), os.ModePerm)
	if err != nil {
		fmt.Println("Failed to save!")
		panic(err)
	}
	fmt.Println("You can render the dot file using Dataviz:")
	fmt.Println("  sfdp -x -Goverlap=scale -Tpng", *outFile, "> output.png")
	fmt.Println("Done")
}
