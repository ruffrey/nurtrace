package main

import (
	"errors"
	"fmt"
	"io/ioutil"
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
		{
			Name:        "train",
			Usage:       "Train a neural network to percept and predict",
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
					Name:  "model, m",
					Usage: "Which perception model to use? charrnn, category",
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
				required := []string{"model", "network", "data", "vocab"}
				for _, field := range required {
					if c.String(field) == "" {
						return errors.New("Missing required argument " + field)
					}
				}
				return nil
			},
			Action: func(c *cli.Context) (err error) {
				// collect arguments and provide defaults
				perceptionModel := c.String("model")
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
					return cmd.Train(perceptionModel, networkInputFile, networkSaveFile, vocabSaveFile, testDataFile, doProfile, initialNetworkNeurons)
				}

				for i := 0; i < iterations; i++ {
					fmt.Println("------ Start Iteration", i+1, "------")
					err = cmd.Train(perceptionModel, networkInputFile, networkSaveFile, vocabSaveFile, testDataFile, doProfile, initialNetworkNeurons)
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
					Name:  "model, m",
					Usage: "Which perception model to use? charrnn, category",
				},
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
			},
			Before: func(c *cli.Context) error {
				if c.Args().First() == "" {
					return errors.New("Missing network filename")
				}
				// validations
				required := []string{"model", "vocab"}
				for _, field := range required {
					if c.String(field) == "" {
						return errors.New("Missing required argument " + field)
					}
				}
				if c.String("seed") == "" && c.String("seed-file") == "" {
					return errors.New("Either --seed or --seed-file is required")
				}
				return nil
			},
			Action: func(c *cli.Context) (err error) {
				networkSaveFile := c.Args().First()
				perceptionModel := c.String("model")
				vocabSaveFile := c.String("vocab")
				seed := []byte(c.String("seed"))
				seedFile := c.String("seed-file")
				if seedFile != "" {
					seed, err = ioutil.ReadFile(seedFile)
					if err != nil {
						return err
					}
				}

				return cmd.Sample(perceptionModel, networkSaveFile, vocabSaveFile, seed)
			},
		},
	}

	app.Run(os.Args)

}
