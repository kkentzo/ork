package main

import (
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

const (
	DEFAULT_ORKFILE = "Orkfile.yml"
)

type Global struct {
	Default string `yaml:"default"`
	Env     Env    `yaml:"env"`
}

type Orkfile struct {
	Global Global  `yaml:"global"`
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

	// populate the task inventory
	f.inventory = Inventory{}
	for _, t := range f.Tasks {
		if err := f.inventory.Add(t); err != nil {
			return err
		}
	}
	return nil
}

// run the requested task
func (f *Orkfile) Run(label string, logger Logger) error {
	task := f.inventory.Find(label)
	if task == nil {
		return fmt.Errorf("task %s does not exist", label)
	}
	return task.Execute(f.Env(), f.inventory, logger)
}

// run the default task (if any)
func (f *Orkfile) RunDefault(logger Logger) error {
	task := f.DefaultTask()
	if task == nil {
		return errors.New("default task has been requested but has not been set")
	}
	return task.Execute(f.Env(), f.inventory, logger)
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

// return the default task or nil if it does not exist
func (f *Orkfile) DefaultTask() *Task {
	return f.inventory.Find(f.Global.Default)
}

func (f *Orkfile) Env() Env {
	return f.Global.Env
}
