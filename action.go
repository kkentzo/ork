package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/google/shlex"
)

type Action struct {
	statement string
	stdin     io.Reader
	logger    Logger
	expandEnv bool
}

func NewAction(statement string) *Action {
	return &Action{
		statement: statement,
		stdin:     os.Stdin,
		expandEnv: true,
	}
}

func (a *Action) WithEnv(env map[string]string) *Action {
	// setup the environment
	for k, v := range env {
		os.Setenv(k, v)
	}
	return a
}

func (a *Action) WithStdin(stdin io.Reader) *Action {
	a.stdin = stdin
	return a
}

func (a *Action) WithLogger(logger Logger) *Action {
	a.logger = logger
	return a
}

func (a *Action) WithEnvExpansion(expandEnv bool) *Action {
	a.expandEnv = expandEnv
	return a
}

func (a *Action) Execute() error {
	if a.expandEnv {
		a.statement = os.ExpandEnv(a.statement)
	}
	cmd, err := createCommand(a.statement)
	if err != nil {
		return err
	}

	cmd.Stderr = os.Stderr
	cmd.Stdin = a.stdin

	var stdout io.ReadCloser
	if stdout, err = cmd.StdoutPipe(); err != nil {
		return fmt.Errorf("failed to connect to shell's standard output: %v", err)
	}

	// spawn the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start the shell process: %v", err)
	}

	// start capturing the shell's stdout
	captureErr := make(chan error, 1)
	go captureAndLogOutput(stdout, a.logger, captureErr)

	// wait for the command to finish
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("process failed: %v", err)
	}

	// wait for command output to finish
	if err := <-captureErr; err != nil {
		return fmt.Errorf("failed to read shell's standard output: %v", err)
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

func captureAndLogOutput(stdout io.ReadCloser, logger Logger, captureErr chan error) {
	// start reading from stdout
	var (
		n   int
		err error
	)
	buf := make([]byte, 256)
	reader := bufio.NewReader(stdout)
	for {
		n, err = reader.Read(buf)
		if n > 0 {
			logger.Output(string(buf[:n]))
		}
		if err != nil {
			if err == io.EOF {
				break
			} else if strings.HasSuffix(err.Error(), "file already closed") {
				// this means that the process has finished and stdout is closed
				break
			} else {
				fmt.Printf("err=%+v\n", err)
				captureErr <- err
				return
			}
		}
	}
	close(captureErr)
}
