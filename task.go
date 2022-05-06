package main

import (
	"fmt"
	"os"
	"os/exec"
)

type Task struct {
	name        string
	description string
	actions     []string
	env         []string
	shell       string
	deps        []*Task
	logger      Logger
}

func NewTask(name string, actions []string, shell string, env []string, logger Logger) *Task {
	return &Task{name: name, actions: actions, shell: shell, env: env, logger: logger}
}

func (t *Task) Info() string {
	var desc string
	if t.description == "" {
		desc = "<no description>"
	} else {
		desc = t.description
	}
	return fmt.Sprintf("%s: %s", t.name, desc)
}

// add another task as a dependency to this task
func (t *Task) AddDependency(other *Task) {
	t.logger.Debugf("adding %s as a dependency of task %s", other.name, t.name)
	t.deps = append(t.deps, other)
}

// execute the task
func (t *Task) Execute() error {
	return t.execute(map[string]struct{}{})
}

func (t *Task) execute(cdt map[string]struct{}) error {
	// mark task as visited
	cdt[t.name] = struct{}{}

	// first, execute all dependencies
	for _, dep := range t.deps {
		// should we visit the dependency?
		if _, ok := cdt[dep.name]; ok {
			return fmt.Errorf("cyclic dependency detected: %s->%s", t.name, dep.name)
		}
		if err := dep.execute(cdt); err != nil {
			return err
		}
	}
	// execute task
	return t.executeActions()
}

// execute the task's actions
// stop execution in case of an error in an action
// return the first encountered error (if any)
func (t *Task) executeActions() error {
	// execute the present task
	for _, action := range t.actions {
		t.logger.Infof("[%s] %s", t.name, action)
		out, err := execute(t.shell, action, t.env)
		t.logger.Output(out)
		if err != nil {
			return err
		}
	}
	return nil
}

// execute the `statement` in `shell` and return the output and/or the error (if any)
// `env` contains "KEY=VAL" items that will be injected into the environment sequentially
func execute(shell string, statement string, env []string) (string, error) {
	// setup command
	cmd := exec.Command(shell, "-c", statement)
	// setup command environment
	cmd.Env = os.Environ()
	for _, kv := range env {
		cmd.Env = append(cmd.Env, kv)
	}
	// run!
	out, err := cmd.Output()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}
