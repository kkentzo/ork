package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	DEFAULT_SHELL   = "/bin/bash"
	DEFAULT_ORKFILE = "./Orkfile.yml"
)

type Global struct {
	Default string            `yaml:"default"`
	Env     map[string]string `yaml:"env"`
	Shell   string            `yaml:"shell"`
}

type OrkfileTask struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Env         map[string]string `yaml:"env"`
	Actions     []string          `yaml:"actions"`
	DependsOn   []string          `yaml:"depends_on"`
}

func (ot OrkfileTask) ToTask(shell string, env []string, logger Logger) *Task {
	return &Task{
		name:        ot.Name,
		description: ot.Description,
		actions:     ot.Actions,
		shell:       shell,
		env:         env,
		logger:      logger,
	}
}

type Orkfile struct {
	Global Global        `yaml:"global"`
	Tasks  []OrkfileTask `yaml:"tasks"`

	path string

	tasks map[string]*Task
}

func Read(path string) (contents []byte, err error) {
	contents, err = ioutil.ReadFile(path)
	if err != nil {
		return
	}
	return
}

func New() *Orkfile {
	return &Orkfile{}
}

func (f *Orkfile) Parse(contents []byte, logger Logger) error {
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
		f.tasks[t.Name] = t.ToTask(f.Global.Shell, mergeEnv(f.Global.Env, t.Env), logger)
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

func (f *Orkfile) AllTasks() []*Task {
	tasks := make([]*Task, 0, len(f.tasks))
	for _, task := range f.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// return the task corresponding to label or nil if it does not exist
func (f *Orkfile) Task(label string) *Task {
	return f.tasks[label]
}

// return the default task or nil if it does not exist
func (f *Orkfile) DefaultTask() *Task {
	return f.tasks[f.Global.Default]
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
