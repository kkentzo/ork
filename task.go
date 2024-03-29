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

type Requirements struct {
	Exists []string          `yaml:"exists"`
	Equals map[string]string `yaml:"equals"`
}

type Task struct {
	label          string
	Name           string        `yaml:"name"`
	Default        string        `yaml:"default"` // used in the global task
	Description    string        `yaml:"description"`
	WorkingDir     string        `yaml:"working_dir"`
	Env            []Env         `yaml:"env"`
	ExpandEnv      *bool         `yaml:"expand_env"`
	GreedyEnvSubst *bool         `yaml:"env_subst_greedy"`
	Actions        []string      `yaml:"actions"`
	DependsOn      []string      `yaml:"depends_on"`
	Tasks          []*Task       `yaml:"tasks"`
	OnSuccess      []string      `yaml:"on_success"`
	OnFailure      []string      `yaml:"on_failure"`
	DynamicTasks   []*Task       `yaml:"generate"`
	Requirements   *Requirements `yaml:"require"`
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
	if parent := findParent(lt.label, inventory); parent != nil {
		if err := parent.WithStdin(lt.stdin).execute(ctx, inventory, logger, cdt); err != nil {
			return err
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

	// are the requirements satisfied?
	if err := lt.CheckRequirements(); err != nil {
		return fmt.Errorf("[%s] failed requirement: %v", lt.label, err)
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

func (lt *LabeledTask) CheckRequirements() error {
	if lt.Requirements == nil {
		return nil
	}
	for _, req := range lt.Requirements.Exists {
		if _, exists := os.LookupEnv(req); !exists {
			return fmt.Errorf("variable %s is not defined ", req)
		}
	}
	for key, expected := range lt.Requirements.Equals {
		actual, exists := os.LookupEnv(key)
		if !exists {
			return fmt.Errorf("variable %s has an expected value but does not exist in the environment", key)
		}
		// expand any environment variables in expected value
		expected = os.ExpandEnv(expected)
		if actual != expected {
			return fmt.Errorf("variable %s exists but does not match its expected value", key)
		}
	}
	return nil
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

// find and return the first parent of the current task if any
// return nil if no parent was found
func findParent(current string, inventory Inventory) *LabeledTask {
	tokens := strings.Split(current, DEFAULT_TASK_GROUP_SEP)
	n := len(tokens) - 1
	for i := n; i > 0; i-- {
		label := strings.Join(tokens[:i], DEFAULT_TASK_GROUP_SEP)
		if parent := inventory.Find(label); parent != nil {
			return parent
		}
	}
	return nil
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
