package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/google/shlex"
)

type Action struct {
	statement string
	env       Env
	stdin     io.Reader
	stdout    io.Writer
	logger    Logger
	expandEnv bool
}

func NewAction(statement string, env Env) *Action {
	return &Action{
		statement: statement,
		env:       env,
		stdin:     os.Stdin,
		stdout:    os.Stdout,
		expandEnv: true,
	}
}

func (a *Action) WithStdin(stdin io.Reader) *Action {
	a.stdin = stdin
	return a
}

func (a *Action) WithStdout(stdout io.Writer) *Action {
	a.stdout = stdout
	return a
}

func (a *Action) WithEnvExpansion(expandEnv bool) *Action {
	a.expandEnv = expandEnv
	return a
}

func (a *Action) Execute() error {
	// first, setup the environment
	if err := a.env.Apply(); err != nil {
		return fmt.Errorf("failed to apply environment: %v", err)
	}

	if a.expandEnv {
		a.statement = os.ExpandEnv(a.statement)
	}
	cmd, err := createCommand(a.statement)
	if err != nil {
		return err
	}

	// setup the process' IO streams
	cmd.Stderr = os.Stderr
	cmd.Stdin = a.stdin
	cmd.Stdout = a.stdout

	// spawn the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start the shell process: %v", err)
	}

	// wait for the command to finish
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("process failed: %v", err)
	}

	return nil
}

func createCommand(statement string) (*exec.Cmd, error) {
	var name string
	var args []string

	fields, err := shlex.Split(statement)
	if err != nil {
		return nil, fmt.Errorf("failed to parse action: %s\nerror: %v", statement, err)
	}
	if len(fields) > 0 {
		name = fields[0]
		args = fields[1:]
	}

	return exec.Command(name, args...), nil
}
