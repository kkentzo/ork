package main

import (
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

type OrkfileTask struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Env         Env      `yaml:"env"`
	Action      string   `yaml:"action"`
	Actions     []string `yaml:"actions"`
	DependsOn   []string `yaml:"depends_on"`
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
	// create all tasks
	for _, t := range f.Tasks {
		if _, ok := f.tasks[t.Name]; ok {
			return fmt.Errorf("duplicate task: %s", t.Name)
		}
		env := f.Global.Env.Copy().Merge(t.Env)
		task, err := NewTask(t, env, logger)
		if err != nil {
			return fmt.Errorf("task: %s: %v", t.Name, err)
		}
		f.tasks[t.Name] = task
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
