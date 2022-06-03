package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// spawn the supplied shell and stream the given statement to the shell process
// capture and return the output per statement
func Execute(shell string, statement string, logger Logger) error {
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
		if err == io.EOF {
			break
		}
		if err != nil {
			captureErr <- err
			return
		}
	}
	captureErr <- nil
}
