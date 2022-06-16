package main

import (
	"fmt"
)

type Task struct {
	name        string
	description string
	env         Env
	actions     []string
	chdir       string
	deps        []*Task
}

func NewTask(ot OrkfileTask) *Task {
	return &Task{
		name:        ot.Name,
		description: ot.Description,
		env:         ot.Env,
		actions:     ot.Actions,
		chdir:       ot.WorkingDir,
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
	t.deps = append(t.deps, other)
}

// execute the task
func (t *Task) Execute(env Env, logger Logger) error {
	return t.execute(env, logger, map[string]struct{}{})
}

func (t *Task) execute(env Env, logger Logger, cdt map[string]struct{}) error {
	// mark task as visited
	cdt[t.name] = struct{}{}

	// first, execute all dependencies
	for _, dep := range t.deps {
		// should we visit the dependency?
		if _, ok := cdt[dep.name]; ok {
			return fmt.Errorf("cyclic dependency detected: %s->%s", t.name, dep.name)
		}
		if err := dep.execute(env, logger, cdt); err != nil {
			return err
		}
	}
	// execute task
	return t.executeActions(env, logger)
}

// execute the task's actions and and log the outputs
// stop execution in case of an error in an action
// return the first encountered error (if any)
func (t *Task) executeActions(env Env, logger Logger) error {
	env = env.Copy().Merge(t.env)

	// execute all the actions
	for idx, action := range t.actions {
		logger.Infof("[%s] %s", t.name, t.actions[idx])
		a := NewAction(action, env).WithStdout(logger).WithWorkingDirectory(t.chdir)
		if err := a.Execute(); err != nil {
			return err
		}
	}
	return nil
}
