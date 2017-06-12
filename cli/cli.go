package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	cmd "github.com/ruffrey/nurtrace/cli/cmd"
	"github.com/ruffrey/nurtrace/potential"
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
	app.Version = "0.13.0"
	app.Copyright = "Symbolic Logic and Jeff Parrish (c) 2017"

	app.EnableBashCompletion = true

	app.Commands = []cli.Command{
		{
			Name:        "train",
			Usage:       "Train a neural network to perceive",
			Description: "Will create the network first, if it does not exist.",
			ArgsUsage:   "",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "network, n",
					Usage: "Network input/output save file",
				},
				cli.StringFlag{
					Name:  "output, o",
					Usage: "Optional network output file if you want it different than --network",
				},
				cli.StringFlag{
					Name:  "vocab, v",
					Usage: "File for loading and saving the vocab",
				},
				cli.StringFlag{
					Name:  "data, d",
					Usage: "Training data file",
				},
				cli.StringFlag{
					Name:  "profile, p",
					Usage: "Optionally train with either 'cpu' or 'mem' profiling enabled, for detecting performance issues and leaks",
				},
				cli.IntFlag{
					Name:  "size, s",
					Usage: "Optionally specify the initial network size when creating it",
				},
				cli.IntFlag{
					Name:  "iterations, i",
					Usage: "Optionally specify the number of times to train on the dataset.",
				},
			},
			Before: func(c *cli.Context) error {
				// validations
				required := []string{"network", "data", "vocab"}
				for _, field := range required {
					if c.String(field) == "" {
						return errors.New("Missing required argument " + field)
					}
				}
				return nil
			},
			Action: func(c *cli.Context) (err error) {
				// collect arguments and provide defaults
				networkInputFile := c.String("network")
				networkSaveFile := c.String("output")
				if networkSaveFile == "" {
					networkSaveFile = networkInputFile
				}
				vocabSaveFile := c.String("vocab")
				testDataFile := c.String("data")
				doProfile := c.String("profile")
				initialNetworkNeurons := c.Int("size")
				if initialNetworkNeurons == 0 {
					initialNetworkNeurons = 200
				}
				iterations := c.Int("iterations")
				if iterations == 0 {
					iterations = 1
				}

				// run it

				if iterations == 1 {
					return cmd.Train(networkInputFile, networkSaveFile, vocabSaveFile, testDataFile, doProfile, initialNetworkNeurons)
				}
				for i := 0; i < iterations; i++ {
					fmt.Println("------ Start Iteration", i+1, "------")
					err = cmd.Train(networkInputFile, networkSaveFile, vocabSaveFile, testDataFile, doProfile, initialNetworkNeurons)
					fmt.Println("------ End Iteration", i+1, "------")
					if err != nil {
						fmt.Println("Failed on iteration", i+1)
						return err
					}
				}

				return err
			},
		},
		{
			Name:      "sample",
			Usage:     "Activate a network to produce a sample (prediction)",
			ArgsUsage: "[network file]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "vocab, v",
					Usage: "File for loading or saving the vocab",
				},
				cli.StringFlag{
					Name:  "seed, s",
					Usage: "Seed the network with data before sampling (text only), to activate network for prediction",
				},
				cli.StringFlag{
					Name:  "seed-file, f",
					Usage: "Read raw seed data from a file, to activate network for prediction",
				},
				cli.IntFlag{
					Name:  "length, l",
					Usage: "Optional length of response to wait for, defaults to 10",
				},
			},
			Before: func(c *cli.Context) error {
				if c.Args().First() == "" {
					return errors.New("Missing network filename")
				}
				// validations
				if c.String("vocab") == "" {
					return errors.New("Missing required argument vocab")
				}
				if c.String("seed") == "" && c.String("seed-file") == "" {
					return errors.New("Either --seed or --seed-file is required")
				}
				return nil
			},
			Action: func(c *cli.Context) (err error) {
				networkSaveFile := c.Args().First()
				vocabSaveFile := c.String("vocab")
				seed := c.String("seed")
				seedFile := c.String("seed-file")
				desiredLength := c.Int("length")
				if seedFile != "" {
					_seed, err := ioutil.ReadFile(seedFile)
					if err != nil {
						return err
					}
					seed = string(_seed)
				}
				if desiredLength >= 0 {
					desiredLength = 10
				}

				return cmd.Sample(networkSaveFile, vocabSaveFile, seed, desiredLength)
			},
		},
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
				cli.BoolFlag{
					Name:  "tags, g",
					Usage: "Print all cells that have a Tag property",
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
				return cmd.Inspect(net, c.Bool("integrity"), c.Bool("totals"), c.Bool("tags"), c.Int("cell"), c.Int("synapse"))
			},
		},
		{
			Name:      "fire",
			Usage:     "Fire a cell and print the firing pattern",
			ArgsUsage: "[network file] [cell ID]",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "n",
					Usage: "Number of times to fire",
				},
			},
			Before: func(c *cli.Context) error {
				if c.Args().First() == "" {
					return errors.New("Missing network filename")
				}
				if c.Args().Get(1) == "" {
					return errors.New("Missing cell to fire")
				}
				return nil
			},
			Action: func(c *cli.Context) (err error) {
				net := c.Args().First()
				cellString := c.Args().Get(1)
				n := c.Int("n")
				var cell potential.CellID
				if n == 0 {
					n = 1
				}
				network, err := potential.LoadNetworkFromFile(net)
				if err != nil {
					return err
				}
				cellInt, err := strconv.Atoi(cellString)
				if err != nil {
					return err
				}
				cell = potential.CellID(cellInt)

				return cmd.FireCell(network, cell, n)
			},
		},
		{
			Name:      "diff-firings",
			Usage:     "Print the difference between the firing pattern of two cell groups. Random cells chosen otherwise ",
			ArgsUsage: "[network file] [cell1 IDs] [cell2 IDs]",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "i",
					Usage: "Number of random cells per group",
				},
				cli.IntFlag{
					Name:  "n",
					Usage: "Number of times to fire",
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
				cell1String := c.Args().Get(1)
				cell2String := c.Args().Get(2)
				r := c.Int("i")
				n := c.Int("n")
				cell1 := make(potential.FiringPattern)
				cell2 := make(potential.FiringPattern)
				if r == 0 {
					r = 4
				}
				if n == 0 {
					n = 1
				}

				network, err := potential.LoadNetworkFromFile(net)
				if err != nil {
					return err
				}

				if cell1String == "" {
					for i := 0; i < r; i++ {
						cell1[network.RandomCellKey()] = 1
					}
				} else {
					cellInts := strings.Split(cell1String, ",")
					for i := 0; i < len(cellInts); i++ {
						cellInt, err := strconv.Atoi(cellInts[i])
						if err != nil {
							return err
						}
						cell1[potential.CellID(cellInt)] = 1
					}

				}
				if cell2String == "" {
					for i := 0; i < r; i++ {
						cell2[network.RandomCellKey()] = 1
					}
				} else {
					cellInts := strings.Split(cell2String, ",")
					for i := 0; i < len(cellInts); i++ {
						cellInt, err := strconv.Atoi(cellInts[i])
						if err != nil {
							return err
						}
						cell2[potential.CellID(cellInt)] = 1
					}

				}

				return cmd.CompareFiringPatterns(network, cell1, cell2, n)
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
