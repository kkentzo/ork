package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	DEFAULT_SHELL   = "/bin/bash"
	DEFAULT_ORKFILE = "Orkfile.yml"
)

type Global struct {
	Default string            `yaml:"default"`
	Env     map[string]string `yaml:"env"`
	Shell   string            `yaml:"shell"`
}

type OrkfileTask struct {
	Name      string            `yaml:"name"`
	Env       map[string]string `yaml:"env"`
	Actions   []string          `yaml:"actions"`
	DependsOn []string          `yaml:"depends_on"`
}

type Orkfile struct {
	Global Global        `yaml:"global"`
	Tasks  []OrkfileTask `yaml:"tasks"`

	path   string
	logger Logger

	tasks map[string]*Task
}

func Read(path string) (contents []byte, err error) {
	contents, err = ioutil.ReadFile(path)
	if err != nil {
		return
	}
	return
}

func New(logger Logger) *Orkfile {
	return &Orkfile{
		logger: logger,
	}
}

func (f *Orkfile) Parse(contents []byte) error {
	f.tasks = map[string]*Task{}

	if err := yaml.Unmarshal(contents, f); err != nil {
		return err
	}
	// determine shell
	if f.Global.Shell == "" {
		f.Global.Shell = pathToShell()
	}
	// create all tasks
	for _, t := range f.Tasks {
		f.tasks[t.Name] = NewTask(t.Name, t.Actions, f.Global.Shell, mergeEnv(f.Global.Env, t.Env), f.logger)
	}
	// create task dependencies
	for _, t := range f.Tasks {
		task := f.tasks[t.Name]
		for _, d := range t.DependsOn {
			if dtask, ok := f.tasks[d]; ok {
				task.AddDependency(dtask)
			} else {
				return fmt.Errorf("task %s (dependency of task %s) does not exist", d, t.Name)
			}
		}
	}
	return nil
}

func (f *Orkfile) Execute(lbl string) error {
	if task, ok := f.tasks[lbl]; ok {
		return task.Execute()

	}
	return fmt.Errorf("task %s does not exist", lbl)
}

// execute the default task (if it exists)
func (f *Orkfile) ExecuteDefault() error {
	if f.Global.Default == "" {
		return errors.New("No default task found in the global section")
	}
	return f.Execute(f.Global.Default)
}

// will merge the local envs into the global ones and return as a list of "KEY=VAL" items
// no de-duplication happens
func mergeEnv(global, local map[string]string) []string {
	env := []string{}
	for k, v := range global {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range local {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

func pathToShell() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	return DEFAULT_SHELL
}
