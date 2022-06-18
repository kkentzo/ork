package main

import (
	"errors"
	"fmt"
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

	inventory Inventory
}

func Read(path string) (contents []byte, err error) {
	contents, err = ioutil.ReadFile(path)
	if err != nil {
		return
	}
	return
}

func New() *Orkfile { return &Orkfile{} }

// parse the orkfile and populate the task inventory
func (f *Orkfile) Parse(contents []byte) error {
	if err := yaml.Unmarshal(contents, f); err != nil {
		return err
	}
	// name the global task (if not already named)
	if f.Global != nil && f.Global.Name == "" {
		f.Global.Name = "global"
	}

	// populate the task inventory
	f.inventory = Inventory{}
	return f.inventory.Populate(f.Tasks)
}

// run the requested task
func (f *Orkfile) Run(label string, logger Logger) error {
	if f.Global != nil {
		if err := f.Global.Execute(f.inventory, logger); err != nil {
			return fmt.Errorf("failed to execute global task: %v", err)
		}
	}
	return run(label, f.inventory, logger)
}

// run the default task (if any)
func (f *Orkfile) RunDefault(logger Logger) error {
	if f.Global == nil || f.Global.Default == "" {
		return errors.New("default task has not been set")
	}
	return f.Run(f.Global.Default, logger)
}

// return info for the requested task
func (f *Orkfile) Info(label string) (info string) {
	if task := f.inventory.Find(label); task != nil {
		info = task.Info()
	}
	return
}

func (f *Orkfile) AllTasks() []*Task {
	return f.inventory.All()
}

func (f *Orkfile) Env() []Env {
	return f.Global.Env
}

func run(label string, inventory Inventory, logger Logger) error {
	task := inventory.Find(label)
	if task == nil {
		return fmt.Errorf("task %s does not exist", label)
	}
	tokens := strings.Split(label, DEFAULT_TASK_GROUP_SEP)
	parentTaskLabel := strings.Join(tokens[0:len(tokens)-1], DEFAULT_TASK_GROUP_SEP)
	if parentTaskLabel != "" {
		if err := run(parentTaskLabel, inventory, logger); err != nil {
			return err
		}
	}
	return task.Execute(inventory, logger)
}
