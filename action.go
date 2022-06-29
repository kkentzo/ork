package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/google/shlex"
)

type Action struct {
	statement string
	chdir     string
	stdin     io.Reader
	stdout    io.Writer
	ctx       context.Context
	logger    Logger
	expandEnv bool
}

func NewAction(ctx context.Context, statement string) *Action {
	return &Action{
		statement: statement,
		stdin:     os.Stdin,
		stdout:    os.Stdout,
		expandEnv: true,
		ctx:       ctx,
	}
}

func (a *Action) WithStdin(stdin io.Reader) *Action {
	if stdin != nil {
		a.stdin = stdin
	}
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

func (a *Action) WithWorkingDirectory(chdir string) *Action {
	a.chdir = chdir
	return a
}

func (a *Action) Execute() error {
	// first, setup the environment
	if a.expandEnv {
		a.statement = os.ExpandEnv(a.statement)
	}
	cmd, err := createCommand(a.ctx, a.statement)
	if err != nil {
		return err
	}

	// setup the process' working directory
	cmd.Dir = a.chdir
	// setup the process' IO streams
	cmd.Stderr = os.Stderr
	cmd.Stdin = a.stdin
	cmd.Stdout = a.stdout

	// spawn the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start action: %v", err)
	}

	// wait for the command to finish
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("action failed: %v", err)
	}

	return nil
}

func createCommand(ctx context.Context, statement string) (*exec.Cmd, error) {
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
	return exec.CommandContext(ctx, name, args...), nil
}
