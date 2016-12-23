package main

import (
	"bleh/potential"
	"flag"
	"fmt"
)

var network = flag.String("n", "", "network filepath - required - Get basic total info about a network.")
var synapse = flag.Uint("s", 0, "synapse ID - Pass a synapse ID to get info about it.")
var cell = flag.Uint("c", 0, "cell ID - Pass a cell ID to get info about it.")
var diff = flag.String("d", "",
	"diff - Pass a second network filepath to calculate the diff using n as the original and d as the forked network.")
var integrity = flag.String("i", "", "integrity - Check network integrity. (-i=ok) or (-i=report)")

func main() {
	flag.Parse()
	var net *potential.Network
	var err error

	if *network != "" {
		net, err = potential.LoadNetworkFromFile(*network)
		if err != nil {
			fmt.Println(err)
			return
		}
		if *integrity == "" && *synapse == 0 && *cell == 0 && *diff == "" {
			net.PrintTotals()
			return
		}
	}
	if *integrity != "" {
		ok, report := potential.CheckIntegrity(net)
		if *integrity == "ok" {
			fmt.Println(ok)
			return
		}
		if *integrity == "report" {
			report.Print()
			return
		}
		fmt.Println("expected -i=ok or -i=report")
		return
	}
	if *diff != "" {
		net2, err := potential.LoadNetworkFromFile(*diff)
		if err != nil {
			fmt.Println(err)
			return
		}
		d := potential.DiffNetworks(net, net2)
		d.Print()
		return
	}
	if *cell != 0 {
		c := net.Cells[potential.CellID(*cell)]
		fmt.Println(c)
		return
	}
	if *synapse != 0 {
		s := net.Synapses[potential.SynapseID(*synapse)]
		fmt.Println(s)
		return
	}

	flag.Usage()
}
