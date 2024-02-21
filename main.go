package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"sort"

	"github.com/urfave/cli/v2"
)

// these will be populated at build time using an ldflag
var (
	GitCommit  string
	OrkVersion string
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

// return all the labels in the orkfile in alphabetical order
func AllLabels(f *Orkfile) []string {
	labels := f.Labels(Actionable)
	sort.Slice(labels, func(i, j int) bool {
		return labels[i] < labels[j]
	})
	return labels
}

func runApp(ctx context.Context, args []string, logger Logger) error {
	app := cli.App{
		Name:        "ork",
		Description: "workflow management for software projects",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f", "p"},
				Usage:   "path to Orkfile",
				Value:   DEFAULT_ORKFILE,
			},
			&cli.StringFlag{
				Name:  "log-level",
				Usage: "log level (one of 'info', 'error', 'debug')",
				Value: LOG_LEVEL_INFO,
			},
			&cli.StringFlag{
				Name:    "search",
				Aliases: []string{"s"},
				Usage:   "print the ork task labels that contain the supplied regex term",
			},
			&cli.BoolFlag{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "list all tasks",
			},
			&cli.BoolFlag{
				Name:    "info",
				Aliases: []string{"i"},
				Usage:   "print information about the provided task",
			},
			&cli.BoolFlag{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "show program version",
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
			if c.Bool("version") {
				fmt.Printf("ork version: %s [%s]\n", OrkVersion, GitCommit)
				os.Exit(0)
			}

			// set log level for logger
			if err := logger.SetLogLevel(c.String("log-level")); err != nil {
				return err
			}

			// read Orkfile contents
			contents, err := Read(c.String("file"))
			if err != nil {
				return fmt.Errorf("failed to find Orkfile in path %s", c.String("file"))
			}
			orkfile := New()
			if err := orkfile.Parse(contents); err != nil {
				return fmt.Errorf("failed to parse Orkfile: %v", err)
			}

			// do we just need to search the labels?
			if c.IsSet("search") {
				term := c.String("search")
				if term == "" {
					return errors.New("no search term provided to -s")
				}
				labels := AllLabels(orkfile)
				for _, label := range labels {
					if matched, err := regexp.MatchString(term, label); err != nil {
						return fmt.Errorf("search term %s is an invalid regular expression", term)
					} else if matched {
						logger.Output(orkfile.Info(label) + "\n")
					}
				}
				return nil
			}

			// read in requested task labels
			labels := c.Args().Slice()

			if c.Bool("list") {
				labels := AllLabels(orkfile)
				for _, label := range labels {
					logger.Output(orkfile.Info(label) + "\n")
				}
			} else {
				return orkfile.Run(ctx, labels, logger)
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

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	go func(cancel context.CancelFunc) {
		// here, we catch the SIGINT signal to prevent the parent process (ork)
		// from being immediately interrupted (and, hence, killed)
		// we rely on the fact that the interrupt signal (SIGINT) will be sent
		// to the entire process group by the shell
		// so it will be received by the current action (os.Process) as well
		<-sigs
		// if this happens, then we need to stop the workflow, so that
		// no more actions are executed
		cancel()
	}(cancel)

	if err := runApp(ctx, os.Args, logger); err != nil {
		logger.Error(err.Error())
	}
}
