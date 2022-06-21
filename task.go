package main

import (
	"fmt"
	"os"
)

type graph map[*Task]*Task

type TaskSelector func(*LabeledTask) bool

var (
	Actionable TaskSelector = func(t *LabeledTask) bool { return t.IsActionable() }
	All        TaskSelector = func(_ *LabeledTask) bool { return true }
)

type LabeledTask struct {
	label string // the task's fully qualified name (the one visible to the user)
	*Task
}

type Task struct {
	label          string
	Name           string   `yaml:"name"`
	Default        string   `yaml:"default"` // used in the global task
	Description    string   `yaml:"description"`
	WorkingDir     string   `yaml:"working_dir"`
	Env            []Env    `yaml:"env"`
	ExpandEnv      *bool    `yaml:"expand_env"`
	GreedyEnvSubst *bool    `yaml:"env_subst_greedy"`
	Actions        []string `yaml:"actions"`
	DependsOn      []string `yaml:"depends_on"`
	Tasks          []*Task  `yaml:"tasks"`
	OnSuccess      []string `yaml:"on_success"`
	OnFailure      []string `yaml:"on_failure"`
	DynamicTasks   []*Task  `yaml:"generate"`
}

// execute the task
func (lt *LabeledTask) Execute(inventory Inventory, logger Logger) error {
	return lt.execute(inventory, logger, graph{})
}

// execute the task workflow
// return the first encountered error (if any)
// cdt records dependencies between tasks in the form: key: parent, value: child
func (lt *LabeledTask) execute(inventory Inventory, logger Logger, cdt graph) (err error) {
	// handle success/failure hooks
	defer func() {
		logger.Debugf("[%s] executing post-action hooks", lt.label)
		var actions []string
		if err == nil {
			actions = lt.OnSuccess
		} else {
			// set the ORK_ERROR env variable
			if os.Setenv("ORK_ERROR", err.Error()) != nil {
				logger.Errorf("[%s] failed to set the ORK_ERROR environment variable", lt.label)
			}
			actions = lt.OnFailure
		}
		for _, a := range actions {
			if err := executeAction(a, lt.ExpandEnv, lt.WorkingDir, logger); err != nil {
				logger.Errorf("[%s] failed to execute hook: %v", lt.label, err)
			}
		}
	}()

	// first, execute all dependencies
	logger.Debugf("[%s] executing dependencies", lt.label)
	for _, label := range lt.DependsOn {
		// find the dependency -- does it exist?
		child := inventory.Find(label)
		if child == nil {
			err = fmt.Errorf("[%s] dependency %s does not exist", lt.label, label)
			return
		}

		// should we visit the dependency?
		if dep := cdt[lt.Task]; child.Task == dep {
			err = fmt.Errorf("[%s] cyclic dependency detected: %s->%s", lt.label, lt.Name, dep.Name)
			return
		}

		// ok, let's run it
		cdt[lt.Task] = child.Task
		if err = child.execute(inventory, logger, cdt); err != nil {
			return
		}
	}

	// apply the environment
	logger.Debugf("[%s] applying task environment", lt.label)
	for _, e := range lt.Env {
		if err = e.Apply(lt.IsEnvSubstGreedy()); err != nil {
			err = fmt.Errorf("[%s] failed to apply environment: %v", lt.label, err)
			return
		}
	}

	// execute all the task's actions (if any)
	logger.Debugf("[%s] executing actions", lt.label)
	for idx, action := range lt.Actions {
		logger.Infof("[%s] %s", lt.label, lt.Actions[idx])
		if err = executeAction(action, lt.ExpandEnv, lt.WorkingDir, logger); err != nil {
			err = fmt.Errorf("[%s] %v", lt.label, err)
			return
		}
	}

	return
}

func (t *Task) IsEnvSubstGreedy() bool {
	if t.GreedyEnvSubst == nil {
		return false
	}
	return *t.GreedyEnvSubst
}

func (t *Task) IsActionable() bool {
	return len(t.Actions) > 0 || len(t.DependsOn) > 0
}

func executeAction(action string, expandEnv *bool, chdir string, logger Logger) error {
	ee := true
	if expandEnv != nil {
		ee = *expandEnv
	}
	a := NewAction(action).WithStdout(logger).WithWorkingDirectory(chdir).WithEnvExpansion(ee)
	if err := a.Execute(); err != nil {
		return err
	}
	return nil

}
