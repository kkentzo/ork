package main

import (
	"fmt"
	"os"
)

type Task struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	WorkingDir  string   `yaml:"working_dir"`
	Env         Env      `yaml:"env"`
	ExpandEnv   *bool    `yaml:"expand_env"`
	Actions     []string `yaml:"actions"`
	DependsOn   []string `yaml:"depends_on"`
	Tasks       []*Task  `yaml:"tasks"`
	OnSuccess   []string `yaml:"on_success"`
	OnFailure   []string `yaml:"on_failure"`
}

func (t *Task) Info() string {
	var desc string
	if t.Description == "" {
		desc = "<no description>"
	} else {
		desc = t.Description
	}
	return fmt.Sprintf("[%s] %s", t.Name, desc)
}

// execute the task
func (t *Task) Execute(env Env, inventory Inventory, logger Logger) error {
	return t.execute(env, inventory, logger, map[string]struct{}{})
}

// execute the task workflow
// return the first encountered error (if any)
func (t *Task) execute(env Env, inventory Inventory, logger Logger, cdt map[string]struct{}) (err error) {
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
			if err := executeAction(a, env, t.ExpandEnv, t.WorkingDir, logger); err != nil {
				logger.Errorf("[%s] failed to execute hook: %v", err)
			}
		}
	}()

	// mark task as visited
	cdt[t.Name] = struct{}{}

	// first, execute all dependencies
	logger.Debugf("[%s] executing dependencies", t.Name)
	for _, label := range t.DependsOn {
		// find the dependency -- does it exist?
		dep := inventory.Find(label)
		if dep == nil {
			err = fmt.Errorf("[%s] dependency %s does not exist", t.Name, label)
			return
		}

		// should we visit the dependency?
		if _, ok := cdt[dep.Name]; ok {
			err = fmt.Errorf("[%s] cyclic dependency detected: %s->%s", t.Name, t.Name, dep.Name)
			return
		}

		// ok, let's run it
		if err = dep.execute(env, inventory, logger, cdt); err != nil {
			return
		}
	}

	// apply the environment
	logger.Debugf("[%s] applying environment", t.Name)
	// first, the outer env
	if err = env.Apply(); err != nil {
		err = fmt.Errorf("[%s] failed to apply environment: %v", t.Name, err)
	}
	// now, the current env
	if err = t.Env.Apply(); err != nil {
		err = fmt.Errorf("[%s] failed to apply environment: %v", t.Name, err)
	}

	// execute all the task's actions (if any)
	logger.Debugf("[%s] executing actions", t.Name)
	for idx, action := range t.Actions {
		logger.Infof("[%s] %s", t.Name, t.Actions[idx])
		if err = executeAction(action, env, t.ExpandEnv, t.WorkingDir, logger); err != nil {
			err = fmt.Errorf("[%s] %v", t.Name, err)
			return
		}
	}

	return
}

func executeAction(action string, env Env, expandEnv *bool, chdir string, logger Logger) error {
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
