package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/ruffrey/nurtrace/potential"
)

// Inspect prints information about a network or requested components of the network.
func Inspect(filename string, integrity bool, totals bool, allTags bool, cell int, synapse int) (err error) {
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

	if allTags {
		for _, c := range net.Cells {
			if c.Tag != "" {
				fmt.Println(c.Tag, c.ID)
			}
		}
		return nil
	}

	if cell != 0 {
		exists := net.CellExists(potential.CellID(cell))
		if !exists {
			return errors.New("Cell " + strconv.Itoa(cell) + "does not exist")
		}
		c := net.Cells[potential.CellID(cell)]
		fmt.Println(c)
		return nil
	}

	if synapse != 0 {
		exists := net.SynExists(potential.SynapseID(synapse))
		if !exists {
			return errors.New("Synapse " + strconv.Itoa(synapse) + "does not exist")
		}
		s := net.Synapses[potential.SynapseID(synapse)]
		fmt.Println(s)
		return nil
	}

	net.Print()
	return nil
}
