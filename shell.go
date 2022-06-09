package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// spawn the supplied shell and stream the given statement to the shell process
// capture and return the output per statement
func Execute(shell string, statement string, env map[string]string, logger Logger, stdin io.Reader) error {
	// setup the environment
	for k, v := range env {
		os.Setenv(k, v)
	}

	cmd := exec.Command(shell, "-c", os.ExpandEnv(statement))
	cmd.Stderr = os.Stderr
	cmd.Stdin = stdin

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
	go captureAndLogOutput(stdout, logger, captureErr)

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
