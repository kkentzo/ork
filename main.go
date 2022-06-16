package main

import (
	"fmt"
	"log"
	"os"

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

func main() {
	logger, err := NewLogger()
	if err != nil {
		fmt.Println("failed to initialize logger")
		os.Exit(1)
	}

	prepareCli()
	app := cli.App{
		Name:        "ork",
		Description: "command workflow management for software projects",
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
			for _, t := range orkfile.Tasks {
				fmt.Println(t.Name)
			}
		},
		Action: func(c *cli.Context) error {
			// set log level for logger
			logger.SetLogLevel(c.String("level"))

			// read Orkfile contents
			contents, err := Read(c.String("path"))
			if err != nil {
				logger.Fatalf("failed to find Orkfile: %v", err)
			}
			orkfile := New()
			if err := orkfile.Parse(contents); err != nil {
				logger.Fatalf("failed to parse Orkfile: %v", err)
			}

			if c.Bool("tasks") {
				fmt.Println("show info for all tasks")
				return nil
			}

			// read in task labels
			labels := c.Args().Slice()

			// if no tasks are requested, then we need to work with the default task (if any)
			if len(labels) == 0 {
				if c.Bool("info") {
					// print info for all tasks
					for _, task := range orkfile.AllTasks() {
						fmt.Println(task.Info())
					}
				} else {
					return orkfile.RunDefault(logger)
				}
				return nil
			}

			// act upon all tasks
			for _, label := range labels {
				if c.Bool("info") {
					fmt.Println(orkfile.Info(label))
				} else if err := orkfile.Run(label, logger); err != nil {
					log.Fatal(err.Error())
				}
			}
			return nil
		},
	}
	err = app.Run(os.Args)
	if err != nil {
		logger.Fatal(err.Error())
	}
}
