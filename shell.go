package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// spawn the supplied shell with the given environment
// and stream the given statements to the shell process
// capture and return the output per statement
func Execute(shell string, env map[string]string, statement string, logger Logger) error {
	cmd := exec.Command(shell)

	// connect to the command's I/O pipes
	var (
		stdin  io.WriteCloser
		stdout io.ReadCloser
		err    error
	)
	if stdin, err = cmd.StdinPipe(); err != nil {
		return fmt.Errorf("failed to connect to shell's standard input: %v", err)
	}
	if stdout, err = cmd.StdoutPipe(); err != nil {
		return fmt.Errorf("failed to connect to shell's standard output: %v", err)
	}

	// start capturing the shell's stdout
	captureErr := make(chan error, 1)
	go captureAndLogOutput(stdout, logger, captureErr)

	// spawn the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start the shell process: %v", err)
	}

	// setup the environment
	for k, v := range env {
		os.Setenv(k, v)
	}

	statement = statement + "\n"
	if _, err := stdin.Write([]byte(os.ExpandEnv(statement))); err != nil {
		return fmt.Errorf("failed to send action to shell: %v", err)
	}
	// we're done -- send EOF to stdin (so that stdout is closed)
	if err := stdin.Close(); err != nil {
		return fmt.Errorf("failed to complete action stream: %v", err)
	}
	// wait for command output to finish
	if err := <-captureErr; err != nil {
		return fmt.Errorf("failed to read shell's standard output: %v", err)
	}
	return nil
}

func captureAndLogOutput(stdout io.ReadCloser, logger Logger, captureErr chan error) {
	// start reading from stdout
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		logger.Output(scanner.Text())
	}
	captureErr <- scanner.Err()
}
