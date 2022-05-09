package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func prepareCli() {
	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.HelpName}} [OPTIONS] [TASK1 TASK2 ...]

OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
`
}

func lookupTasks(orkfile *Orkfile, labels []string) ([]*Task, error) {
	// ok, we're not in the default task case
	// lookup all tasks
	tasks := make([]*Task, 0, len(labels))
	for _, label := range labels {
		if task := orkfile.Task(label); task != nil {
			tasks = append(tasks, task)
		} else {
			return tasks, fmt.Errorf("task %s not found in Orkfile", label)
		}
	}
	return tasks, nil
}

func printTasks(orkfile *Orkfile, labels []string) {
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
			&cli.StringFlag{Name: "path", Aliases: []string{"p"}, Usage: "path to Orkfile", Value: DEFAULT_ORKFILE},
			&cli.BoolFlag{Name: "info", Aliases: []string{"i"}, Usage: "show info for the supplied task or all tasks"},
		},
		Action: func(c *cli.Context) error {
			// read Orkfile contents
			contents, err := Read(c.String("path"))
			if err != nil {
				logger.Fatalf("failed to find Orkfile: %v", err)
			}
			orkfile := New()
			if err := orkfile.Parse(contents, logger); err != nil {
				os.Exit(1)
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
					if task := orkfile.DefaultTask(); task != nil {
						return task.Execute()
					} else {
						return errors.New("no default task was defined in Orkfile")
					}
				}
				return nil
			}

			// lookup the requested tasks
			tasks, err := lookupTasks(orkfile, labels)
			if err != nil {
				logger.Fatal(err.Error())
			}

			// act upon all tasks
			for _, task := range tasks {
				if c.Bool("info") {
					fmt.Println(task.Info())
				} else if err := task.Execute(); err != nil {
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
