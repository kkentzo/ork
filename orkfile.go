package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

const (
	DEFAULT_ORKFILE = "Orkfile.yml"
)

type Orkfile struct {
	Default string  `yaml:"default"`
	Tasks   []*Task `yaml:"tasks"`

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
	// populate the task inventory
	f.inventory = Inventory{}
	return f.inventory.Populate(f.Tasks)
}

// run the requested task
func (f *Orkfile) Run(ctx context.Context, label string, logger Logger) error {
	task := f.inventory.Find(label)
	if task == nil {
		return fmt.Errorf("task %s does not exist", label)
	}

	return task.WithStdin(f.stdin).Execute(ctx, f.inventory, logger)
}

// run the default task (if any)
func (f *Orkfile) RunDefault(ctx context.Context, logger Logger) error {
	if f.Default == "" {
		return errors.New("default task has not been set")
	}
	return f.WithStdin(f.stdin).Run(ctx, f.Default, logger)
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
