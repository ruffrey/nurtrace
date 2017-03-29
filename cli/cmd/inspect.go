package cmd

import (
	"errors"
	"flag"
	"fmt"
	"strconv"

	"github.com/ruffrey/nurtrace/potential"
)

var synapse = flag.Uint("s", 0, "synapse ID - Pass a synapse ID to get info about it.")
var cell = flag.Uint("c", 0, "cell ID - Pass a cell ID to get info about it.")
var integrity = flag.String("i", "", "integrity - Check network integrity. (-i=ok) or (-i=report)")

// Inspect prints information about a network or requested components of the network.
func Inspect(filename string, integrity bool, totals bool, cell int, synapse int) (err error) {
	net, err := potential.LoadNetworkFromFile(filename)
	if err != nil {
		return err
	}

	if integrity {
		ok, report := potential.CheckIntegrity(net)
		if ok {
			fmt.Println("Integrity OK - no dangling connections.")
			return nil
		}
		report.Print()
		return errors.New("Failed network integrity check")
	}

	if totals {
		net.PrintTotals()
		return nil
	}

	if cell != 0 {
		c, exists := net.Cells[potential.CellID(cell)]
		if !exists {
			return errors.New("Cell " + strconv.Itoa(cell) + "does not exist")
		}
		fmt.Println(c)
		return nil
	}

	if synapse != 0 {
		s, exists := net.Synapses[potential.SynapseID(synapse)]
		if !exists {
			return errors.New("Synapse " + strconv.Itoa(synapse) + "does not exist")
		}
		fmt.Println(s)
		return nil
	}

	net.Print()
	return nil
}
