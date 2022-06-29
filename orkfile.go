package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DEFAULT_ORKFILE = "Orkfile.yml"
)

type Orkfile struct {
	Global *Task   `yaml:"global"`
	Tasks  []*Task `yaml:"tasks"`

	global    *LabeledTask
	inventory Inventory
	stdin     io.Reader
}

func Read(path string) (contents []byte, err error) {
	contents, err = ioutil.ReadFile(path)
	if err != nil {
		return
	}
	return
}

func New() *Orkfile { return &Orkfile{} }

func (f *Orkfile) WithStdin(stdin io.Reader) *Orkfile {
	f.stdin = stdin
	return f
}

// parse the orkfile and populate the task inventory
func (f *Orkfile) Parse(contents []byte) error {
	if err := yaml.Unmarshal(contents, f); err != nil {
		return err
	}
	// name the global task (if not already named)
	if f.Global != nil {
		f.global = &LabeledTask{label: "global", Task: f.Global}
	}

	// populate the task inventory
	f.inventory = Inventory{}
	return f.inventory.Populate(f.Tasks)
}

// run the requested task
func (f *Orkfile) Run(ctx context.Context, label string, logger Logger) error {
	if f.global != nil {
		if err := f.global.WithStdin(f.stdin).Execute(ctx, f.inventory, logger); err != nil {
			return fmt.Errorf("failed to execute global task: %v", err)
		}
	}
	return run(ctx, label, f.inventory, logger, f.stdin)
}

// run the default task (if any)
func (f *Orkfile) RunDefault(ctx context.Context, logger Logger) error {
	if f.Global == nil || f.Global.Default == "" {
		return errors.New("default task has not been set")
	}
	return f.WithStdin(f.stdin).Run(ctx, f.Global.Default, logger)
}

// return info for the requested task
func (f *Orkfile) Info(label string) (info string) {
	if task := f.inventory.Find(label); task != nil {
		desc := task.Description
		if desc == "" {
			desc = "<no description>"
		}
		info = fmt.Sprintf("[%s] %s", label, desc)
	}
	return
}

func (f *Orkfile) GetTasks(sel TaskSelector) []*LabeledTask {
	return f.inventory.Tasks(sel)
}

func (f *Orkfile) Labels(sel TaskSelector) []string {
	return f.inventory.Labels(sel)
}

func (f *Orkfile) Env() []Env {
	return f.Global.Env
}

func run(ctx context.Context, label string, inventory Inventory, logger Logger, stdin io.Reader) error {
	task := inventory.Find(label)
	if task == nil {
		return fmt.Errorf("task %s does not exist", label)
	}
	tokens := strings.Split(label, DEFAULT_TASK_GROUP_SEP)
	parentTaskLabel := strings.Join(tokens[0:len(tokens)-1], DEFAULT_TASK_GROUP_SEP)
	if parentTaskLabel != "" {
		if err := run(ctx, parentTaskLabel, inventory, logger, stdin); err != nil {
			return err
		}
	}
	return task.WithStdin(stdin).Execute(ctx, inventory, logger)
}
