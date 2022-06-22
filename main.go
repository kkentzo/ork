package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/urfave/cli/v2"
)

func prepareCli() {
	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Description}}

USAGE:
   {{.HelpName}} [OPTIONS] [TASK1 TASK2 ...]

OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
`
}

func runApp(args []string, logger Logger) error {
	app := cli.App{
		Name:        "ork",
		Description: "workflow management for software projects",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "path",
				Aliases: []string{"p"},
				Usage:   "path to Orkfile",
				Value:   DEFAULT_ORKFILE,
			},
			&cli.StringFlag{
				Name:    "level",
				Aliases: []string{"l"},
				Usage:   "log level (one of 'info', 'error', 'debug')",
				Value:   LOG_LEVEL_INFO,
			},
			&cli.BoolFlag{
				Name:    "info",
				Aliases: []string{"i"},
				Usage:   "show info for the supplied task or all tasks",
			},
		},
		EnableBashCompletion: true,
		BashComplete: func(c *cli.Context) {
			// read Orkfile contents
			contents, err := Read(DEFAULT_ORKFILE)
			if err != nil {
				return
			}
			// parse file
			orkfile := New()
			if err := orkfile.Parse(contents); err != nil {
				return
			}
			// return the available task to `complete` command
			for _, lbl := range orkfile.Labels(Actionable) {
				fmt.Println(lbl)
			}
		},
		Action: func(c *cli.Context) error {
			// set log level for logger
			if err := logger.SetLogLevel(c.String("level")); err != nil {
				return err
			}

			// read Orkfile contents
			contents, err := Read(c.String("path"))
			if err != nil {
				return fmt.Errorf("failed to find Orkfile in path %s", c.String("path"))
			}
			orkfile := New()
			if err := orkfile.Parse(contents); err != nil {
				return fmt.Errorf("failed to parse Orkfile: %v", err)
			}

			// read in task labels
			labels := c.Args().Slice()

			// if no tasks are requested, then we
			// either print info for all tasks
			// or we execute the default task
			if len(labels) == 0 {
				if c.Bool("info") {
					// get tasks and sort them by name
					labels := orkfile.Labels(Actionable)
					sort.Slice(labels, func(i, j int) bool {
						return labels[i] < labels[j]
					})
					for _, label := range labels {
						logger.Output(orkfile.Info(label) + "\n")
					}
				} else {
					return orkfile.RunDefault(logger)
				}
				return nil
			}

			// act upon all tasks
			for _, label := range labels {
				if c.Bool("info") {
					logger.Output(orkfile.Info(label) + "\n")
				} else if err := orkfile.Run(label, logger); err != nil {
					return err
				}
			}
			return nil
		},
	}
	return app.Run(args)
}

func main() {
	prepareCli()

	logger, err := NewLogger()
	if err != nil {
		fmt.Println("failed to initialize logger")
		os.Exit(1)
	}

	if err := runApp(os.Args, logger); err != nil {
		logger.Fatal(err.Error())
	}
}
