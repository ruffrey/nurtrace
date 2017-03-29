package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cmd "github.com/ruffrey/nurtrace/cli/cmd"
	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	cli.BashCompletionFlag = cli.BoolFlag{
		Name:   "compgen",
		Hidden: true,
	}

	app := cli.NewApp()
	app.Name = "nt"
	app.Usage = "Nurtrace - generic neural network library"
	app.UsageText = "nt [global options] command [options]"
	app.HelpName = "nt"
	app.Version = "0.12.0"
	app.Copyright = "Symbolic Logic (c) 2017"

	app.EnableBashCompletion = true

	app.Commands = []cli.Command{
		{
			Name:      "merge",
			Usage:     "Merge a neural network onto another one",
			ArgsUsage: "--from=net1.json --to=net2.json",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "from, f",
					Usage: "Network to merge onto the other",
				},
				cli.StringFlag{
					Name:  "to, t",
					Usage: "Network which will receive the merge",
				},
				cli.StringFlag{
					Name:  "output, o",
					Usage: "File to write the new network into",
				},
				cli.BoolFlag{
					Name:  "diff",
					Usage: "Only show the diff between networks",
				},
				cli.BoolFlag{
					Name:  "stdout",
					Usage: "Write result network to stdout (instead of a file)",
				},
			},
			Before: func(c *cli.Context) error {
				if c.String("from") == "" || c.String("to") == "" {
					return errors.New("Missing network filename param(s)")
				}
				if c.String("output") == "" && !c.BoolT("diff") {
					return errors.New("Specify output file for network or stdout")
				}
				return nil
			},
			Action: func(c *cli.Context) (err error) {
				flagOutput := c.String("output")
				flagDiff := c.Bool("diff")
				flagStdout := c.Bool("stdout")

				network, diff, err := cmd.Merge(c.String("to"), c.String("from"))
				if err != nil {
					return err
				}

				if flagDiff {
					diff.Print()
					return nil
				}

				if flagStdout {
					network.Print()
				}

				err = network.SaveToFile(flagOutput)

				return err
			},
		},
		{
			Name:      "inspect",
			Usage:     "Get information about cells and synapses in a network. Prints the network in human readable format by default.",
			ArgsUsage: "[network file]",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "integrity, i",
					Usage: "Check and report integrity for the network",
				},
				cli.BoolFlag{
					Name:  "totals, t",
					Usage: "Only print the totals about the network, instead of the entire network.",
				},
				cli.IntFlag{
					Name:  "cell, c",
					Usage: "Print info about a specific cell",
				},
				cli.IntFlag{
					Name:  "synapse, s",
					Usage: "Print info about a specific synapse",
				},
			},
			Before: func(c *cli.Context) error {
				if c.Args().First() == "" {
					return errors.New("Missing network filename")
				}
				return nil
			},
			Action: func(c *cli.Context) (err error) {
				net := c.Args().First()
				fmt.Println("Reading network from", net)
				return cmd.Inspect(net, c.Bool("integrity"), c.Bool("totals"), c.Int("cell"), c.Int("synapse"))
			},
		},
		{
			Name:        "export",
			Usage:       "Output a network to a different file format",
			Description: "Valid formats: dot, json, default",
			ArgsUsage:   "[network file] [format]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "output, o",
					Usage: "Optional output file",
				},
			},
			Before: func(c *cli.Context) error {
				if c.Args().First() == "" {
					return errors.New("Missing network filename")
				}
				return nil
			},
			Action: func(c *cli.Context) (err error) {
				basename := c.Args().First()
				outFormat := c.Args().Get(1)
				outFormatName := outFormat
				if outFormatName == "default" {
					outFormatName = "nur"
				}
				outFile := c.String("output")
				if outFile == "" {
					outFile = strings.TrimSuffix(basename, filepath.Ext(basename)) + "." + outFormatName
				}

				return cmd.Export(outFormat, basename, outFile)
			},
		},
	}

	app.Run(os.Args)

}
