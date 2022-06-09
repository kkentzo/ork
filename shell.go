package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Shell struct {
	shell     string
	stdin     io.Reader
	logger    Logger
	expandEnv bool
}

func NewShell(shell string) *Shell {
	return &Shell{
		shell:     shell,
		stdin:     os.Stdin,
		expandEnv: true,
	}
}

func (sh *Shell) WithEnv(env map[string]string) *Shell {
	// setup the environment
	for k, v := range env {
		os.Setenv(k, v)
	}
	return sh
}

func (sh *Shell) WithStdin(stdin io.Reader) *Shell {
	sh.stdin = stdin
	return sh
}

func (sh *Shell) WithLogger(logger Logger) *Shell {
	sh.logger = logger
	return sh
}

// set whether the statement to be executed will be env-expanded
// this means that any form containing $VAR or ${VAR} will be replaced with the value from env
// if this is on, it enables command substitution in env variables
// if this is off, it enables reading env variables that were set during the command's execution
// default value: true
func (sh *Shell) SetExpandEnv(expandEnv bool) {
	sh.expandEnv = expandEnv
}

// spawn the supplied shell and stream the given statement to the shell process
// process output is logged using the shell's logger
func (sh *Shell) Execute(statement string) error {
	if sh.expandEnv {
		statement = os.ExpandEnv(statement)
	}
	cmd := exec.Command(sh.shell, "-c", statement)
	cmd.Stderr = os.Stderr
	cmd.Stdin = sh.stdin

	var (
		stdout io.ReadCloser
		err    error
	)
	if stdout, err = cmd.StdoutPipe(); err != nil {
		return fmt.Errorf("failed to connect to shell's standard output: %v", err)
	}

	// spawn the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start the shell process: %v", err)
	}

	// start capturing the shell's stdout
	captureErr := make(chan error, 1)
	go captureAndLogOutput(stdout, sh.logger, captureErr)

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
