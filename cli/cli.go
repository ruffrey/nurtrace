package main

import (
	"errors"
	"os"

	cmd "github.com/ruffrey/nurtrace/cli/cmd"
	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	cli.BashCompletionFlag = cli.BoolFlag{
		Name:   "compgen",
		Hidden: true,
	}

	app := cli.NewApp()
	app.EnableBashCompletion = true

	app.Commands = []cli.Command{
		{
			Name:  "merge",
			Usage: "merge a neural network onto another one",
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
				cli.BoolTFlag{
					Name:  "diff",
					Usage: "Only show the diff between networks",
				},
				cli.BoolTFlag{
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
				flagDiff := c.BoolT("diff")
				flagStdout := c.BoolT("stdout")

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
	}

	app.Run(os.Args)

}
