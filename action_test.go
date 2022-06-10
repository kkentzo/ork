package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Action_Execute(t *testing.T) {
	emptyenv := map[string]string{}

	kases := []struct {
		statement string
		env       map[string]string
		output    string
	}{
		// action prints output
		{"echo foo\n", emptyenv, "foo\n"},
		// action can see the environment
		{"echo ${A_RANDOM_VAR}", map[string]string{"A_RANDOM_VAR": "foo"}, "foo\n"},
		// // command substitution is supported in statements
		// {DEFAULT_SHELL, "echo $(echo foo)", emptyenv, "foo\n"},
		// // command substitution is supported in the environment variables
		// {DEFAULT_SHELL, "echo ${A_RANDOM_VAR}", map[string]string{"A_RANDOM_VAR": "$(echo foo)"}, "foo\n"},
		{"python -c \"import sys; sys.stdout.write('hello from python');\"", emptyenv, "hello from python"},
	}

	for idx, kase := range kases {
		logger := NewMockLogger()
		action := NewAction(kase.statement).WithEnv(kase.env).WithLogger(logger)
		assert.NoError(t, action.Execute(), fmt.Sprintf("test case: %d", idx))
		assert.Contains(t, logger.Outputs(), kase.output, fmt.Sprintf("test case: %d", idx))
	}
}

func Test_Action_Execute_Errors(t *testing.T) {
	kases := []struct {
		statement string
		env       map[string]string
		errmsg    string
		output    string
	}{
		// command prints error
		{"bash -c \"echo foo; exit 1\"", map[string]string{}, "exit status 1", "foo\n"},
	}

	for _, kase := range kases {
		logger := NewMockLogger()
		action := NewAction(kase.statement).WithEnv(kase.env).WithLogger(logger)
		assert.ErrorContains(t, action.Execute(), kase.errmsg)
		assert.Contains(t, logger.Outputs(), kase.output)
	}
}

func Test_Action_Can_Accept_StandardInput(t *testing.T) {
	logger := NewMockLogger()
	b := bytes.NewBufferString("hello\n")
	// we need to disable env expansion in statement so that `$s` is not replaced
	// with empty space (`$s` will be set during command execution, not before)
	action := NewAction("bash -c \"read s && echo $s\"").WithStdin(b).WithEnvExpansion(false).WithLogger(logger)
	assert.NoError(t, action.Execute())
	assert.Contains(t, logger.Outputs(), "hello\n")
}
