package main

import (
	"fmt"
)

type Task struct {
	name        string
	description string
	actions     []string
	env         map[string]string
	deps        []*Task
	logger      Logger
}

func NewTask(ot OrkfileTask, env map[string]string, logger Logger) *Task {
	return &Task{
		name:        ot.Name,
		description: ot.Description,
		actions:     ot.Actions,
		env:         env,
		logger:      logger,
	}
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

// execute the task's actions and and log the outputs
// stop execution in case of an error in an action
// return the first encountered error (if any)
func (t *Task) executeActions() error {
	// execute all the actions
	for idx, action := range t.actions {
		t.logger.Infof("[%s] %s", t.name, t.actions[idx])
		sh := NewAction(action).WithEnv(t.env).WithStdout(t.logger)
		if err := sh.Execute(); err != nil {
			return err
		}
	}
	return nil
}
