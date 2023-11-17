package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
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
	stdin io.Reader
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

func (lt *LabeledTask) WithStdin(stdin io.Reader) *LabeledTask {
	lt.stdin = stdin
	return lt
}

// execute the task
func (lt *LabeledTask) Execute(ctx context.Context, inventory Inventory, logger Logger) error {
	return lt.execute(ctx, inventory, logger, graph{})
}

// execute the task workflow
// return the first encountered error (if any)
// cdt records dependencies between tasks in the form: key: parent, value: child
func (lt *LabeledTask) execute(ctx context.Context, inventory Inventory, logger Logger, cdt graph) (err error) {
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
			if err := executeAction(ctx, a, lt.ExpandEnv, lt.WorkingDir, logger, lt.stdin); err != nil {
				logger.Errorf("[%s] failed to execute hook: %v", lt.label, err)
			}
		}
	}()

	// let's visit and execute any parent tasks first recursively
	tokens := strings.Split(lt.label, DEFAULT_TASK_GROUP_SEP)
	parentTaskLabel := strings.Join(tokens[0:len(tokens)-1], DEFAULT_TASK_GROUP_SEP)
	if parentTaskLabel != "" {
		// we will accept not finding the parent task because it may not really exist as a separate task
		// this is the case where a user has specified explicitly the task a.b in the Orkfile
		// instead of nesting the task b under task a
		if parent := inventory.Find(parentTaskLabel); parent != nil {
			if err := parent.WithStdin(lt.stdin).execute(ctx, inventory, logger, cdt); err != nil {
				return err
			}
		}
	}

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
		if err = child.execute(ctx, inventory, logger, cdt); err != nil {
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
		if err = executeAction(ctx, action, lt.ExpandEnv, lt.WorkingDir, logger, lt.stdin); err != nil {
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

func executeAction(ctx context.Context, action string, expandEnv *bool, chdir string, logger Logger, stdin io.Reader) error {
	ee := true
	if expandEnv != nil {
		ee = *expandEnv
	}
	a := NewAction(action).WithStdout(logger).WithWorkingDirectory(chdir).WithEnvExpansion(ee).WithStdin(stdin)
	if err := a.Execute(); err != nil {
		return err
	}
	// should we proceed to the next action?
	select {
	case <-ctx.Done():
		return errors.New("C-c received")
	default:
		return nil
	}
}
