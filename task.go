package main

import (
	"fmt"
)

type Task struct {
	name        string
	description string
	env         Env
	actions     []string
	onSuccess   []string
	onFailure   []string
	chdir       string
	deps        []*Task
}

func NewTask(ot OrkfileTask) *Task {
	return &Task{
		name:        ot.Name,
		description: ot.Description,
		env:         ot.Env,
		actions:     ot.Actions,
		onSuccess:   ot.OnSuccess,
		onFailure:   ot.OnFailure,
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

// execute the task workflow
// return the first encountered error (if any)
func (t *Task) execute(env Env, logger Logger, cdt map[string]struct{}) (err error) {
	// handle success/failure hooks
	defer func() {
		var actions []string
		if err == nil {
			actions = t.onSuccess
		} else {
			actions = t.onFailure
		}
		for _, a := range actions {
			if err := executeAction(a, env.Copy(), t.chdir, logger); err != nil {
				logger.Errorf("[%s] failed to execute hook: %v", err)
			}
		}
	}()

	// mark task as visited
	cdt[t.name] = struct{}{}

	// first, execute all dependencies
	for _, dep := range t.deps {
		// should we visit the dependency?
		if _, ok := cdt[dep.name]; ok {
			err = fmt.Errorf("[%s] cyclic dependency detected: %s->%s", t.name, t.name, dep.name)
			return
		}
		if err = dep.execute(env, logger, cdt); err != nil {
			return
		}
	}

	// execute all the actions
	for idx, action := range t.actions {
		logger.Infof("[%s] %s", t.name, t.actions[idx])
		if err = executeAction(action, env.Copy().Merge(t.env), t.chdir, logger); err != nil {
			err = fmt.Errorf("[%s] %v", t.name, err)
			return
		}
	}

	return
}

func executeAction(action string, env Env, chdir string, logger Logger) error {
	a := NewAction(action, env).WithStdout(logger).WithWorkingDirectory(chdir)
	if err := a.Execute(); err != nil {
		return err
	}
	return nil

}
