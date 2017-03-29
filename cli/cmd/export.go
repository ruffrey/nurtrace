package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/ruffrey/nurtrace/potential"

	"github.com/awalterschulze/gographviz"
)

const directed bool = false

// Export takes a network and puts it into the requested format
func Export(outFormat, networkFile, outFile string) (err error) {
	network, err := potential.LoadNetworkFromFile(networkFile)
	if err != nil {
		return err
	}

	switch outFormat {
	case "dot":
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

		fmt.Println("- Writing to", outFile)
		err = ioutil.WriteFile(outFile, []byte(output), os.ModePerm)
		if err != nil {
			fmt.Println("Failed to save!")
			return err
		}
		// fmt.Println("You can render the dot file using Dataviz:")
		// fmt.Println("  sfdp -x -Goverlap=scale -Tpng", *outFile, "> output.png")
		// fmt.Println("Done")
		return nil
	case "json":
		return network.SaveToFileReadable(outFile)
	case "default":
		return network.SaveToFile(outFile)
	}

	return errors.New("Output format " + outFormat + " is not a supported option")
}
