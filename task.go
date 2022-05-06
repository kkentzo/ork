package main

import (
	"fmt"
	"os"
	"os/exec"
)

type TaskRegistry map[string]*Task

func NewTaskRegistry(f *Orkfile) TaskRegistry {
	r := TaskRegistry{}
	shell := pathToShell()
	for _, t := range f.Tasks {
		t.global = f.Global
		t.envvars = mergeEnv(f.Global.Env, t.Env)
		t.shell = shell
		r[t.Name] = t
	}
	return r
}

type Task struct {
	Name      string            `yaml:"name"`
	Env       map[string]string `yaml:"env"`
	Actions   []string          `yaml:"actions"`
	DependsOn []string          `yaml:"depends_on"`
	global    *Global
	envvars   []string
	shell     string
}

func (t *Task) Execute() error {
	for _, action := range t.Actions {
		fmt.Printf("[%s] %s", t.Name, action)
		if _, err := execute(t.shell, action, t.envvars); err != nil {
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

func pathToShell() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	return "/bin/bash"
}

// will merge the local envs into the global ones and return as a list of "KEY=VAL" items
// no de-duplication happens
func mergeEnv(global, local map[string]string) []string {
	env := []string{}
	for k, v := range global {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range local {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}
