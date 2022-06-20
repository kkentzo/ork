package main

import (
	"fmt"
	"os"
)

type graph map[*Task]*Task

type Task struct {
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
func (t *Task) Execute(inventory Inventory, logger Logger) error {
	return t.execute(inventory, logger, graph{})
}

// execute the task workflow
// return the first encountered error (if any)
// cdt records dependencies between tasks in the form: key: parent, value: child
func (t *Task) execute(inventory Inventory, logger Logger, cdt graph) (err error) {
	// handle success/failure hooks
	defer func() {
		logger.Debugf("[%s] executing post-action hooks", t.Name)
		var actions []string
		if err == nil {
			actions = t.OnSuccess
		} else {
			// set the ORK_ERROR env variable
			if os.Setenv("ORK_ERROR", err.Error()) != nil {
				logger.Errorf("[%s] failed to set the ORK_ERROR environment variable", t.Name)
			}
			actions = t.OnFailure
		}
		for _, a := range actions {
			if err := executeAction(a, t.ExpandEnv, t.WorkingDir, logger); err != nil {
				logger.Errorf("[%s] failed to execute hook: %v", t.Name, err)
			}
		}
	}()

	// first, execute all dependencies
	logger.Debugf("[%s] executing dependencies", t.Name)
	for _, label := range t.DependsOn {
		// find the dependency -- does it exist?
		child := inventory.Find(label)
		if child == nil {
			err = fmt.Errorf("[%s] dependency %s does not exist", t.Name, label)
			return
		}

		// should we visit the dependency?
		if dep := cdt[t]; child == dep {
			err = fmt.Errorf("[%s] cyclic dependency detected: %s->%s", t.Name, t.Name, dep.Name)
			return
		}

		// ok, let's run it
		cdt[t] = child
		if err = child.execute(inventory, logger, cdt); err != nil {
			return
		}
	}

	// apply the environment
	logger.Debugf("[%s] applying task environment", t.Name)
	for _, e := range t.Env {
		if err = e.Apply(t.IsEnvSubstGreedy()); err != nil {
			err = fmt.Errorf("[%s] failed to apply environment: %v", t.Name, err)
			return
		}
	}

	// execute all the task's actions (if any)
	logger.Debugf("[%s] executing actions", t.Name)
	for idx, action := range t.Actions {
		logger.Infof("[%s] %s", t.Name, t.Actions[idx])
		if err = executeAction(action, t.ExpandEnv, t.WorkingDir, logger); err != nil {
			err = fmt.Errorf("[%s] %v", t.Name, err)
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
