package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Bash_Shell_Execute(t *testing.T) {
	kases := []struct {
		shell     string
		statement string
		env       map[string]string
		output    string
	}{
		// command prints output
		{"/bin/bash", "echo foo", map[string]string{}, "foo\n"},
		// shell can see the environment
		{"/bin/bash", "echo ${A_RANDOM_VAR}", map[string]string{"A_RANDOM_VAR": "foo"}, "foo\n"},
		// command substitution is supported in statements
		{"/bin/bash", "echo $(echo foo)", map[string]string{}, "foo\n"},
		// command substitution is supported in the environment variables
		{"/bin/bash", "echo ${A_RANDOM_VAR}", map[string]string{"A_RANDOM_VAR": "$(echo foo)"}, "foo\n"},
	}

	for idx, kase := range kases {
		logger := NewMockLogger()
		sh := NewShell(kase.shell).WithEnv(kase.env).WithLogger(logger)
		assert.NoError(t, sh.Execute(kase.statement), fmt.Sprintf("test case: %d", idx))
		assert.Contains(t, logger.Outputs(), kase.output, fmt.Sprintf("test case: %d", idx))
	}
}

func Test_Bash_Shell_Execute_Errors(t *testing.T) {
	kases := []struct {
		shell     string
		statement string
		env       map[string]string
		errmsg    string
		output    string
	}{
		// command prints error
		{"/bin/bash", "echo foo; exit 1", map[string]string{}, "exit status 1", "foo\n"},
	}

	for _, kase := range kases {
		logger := NewMockLogger()
		sh := NewShell(kase.shell).WithEnv(kase.env).WithLogger(logger)
		assert.ErrorContains(t, sh.Execute(kase.statement), kase.errmsg)
		assert.Contains(t, logger.Outputs(), kase.output)
	}
}

func Test_Shell_Stamenent_Can_Accept_StandardInput(t *testing.T) {
	logger := NewMockLogger()
	b := bytes.NewBufferString("hello\n")
	sh := NewShell("/bin/bash").WithLogger(logger).WithStdin(b)
	// we need to disable env expansion in statement so that `$s` is not replaced
	// with empty space (`$s` will be set during command execution, not before)
	sh.SetExpandEnv(false)
	assert.NoError(t, sh.Execute("read s && echo $s"))
	assert.Contains(t, logger.Outputs(), "hello\n")
}
