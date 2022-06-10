package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Action_Execute(t *testing.T) {
	emptyenv := Env{}

	kases := []struct {
		statement string
		env       Env
		output    string
	}{
		// action prints output
		{"echo foo\n", emptyenv, "foo\n"},
		// action can see the environment
		{"echo ${A_RANDOM_VAR}", Env{"A_RANDOM_VAR": "foo"}, "foo\n"},
		// action can execute arbitrary commands
		{"python -c \"import sys; sys.stdout.write('hello from python');\"", emptyenv, "hello from python"},
	}

	for idx, kase := range kases {
		logger := NewMockLogger()
		action := NewAction(kase.statement, kase.env).WithStdout(logger)
		assert.NoError(t, action.Execute(), fmt.Sprintf("test case: %d", idx))
		assert.Contains(t, logger.Outputs(), kase.output, fmt.Sprintf("test case: %d", idx))
	}
}

func Test_Action_Execute_Errors(t *testing.T) {
	kases := []struct {
		statement string
		env       Env
		errmsg    string
		output    string
	}{
		// command prints error
		{"bash -c \"echo foo; exit 1\"", Env{}, "exit status 1", "foo\n"},
	}

	for _, kase := range kases {
		logger := NewMockLogger()
		action := NewAction(kase.statement, kase.env).WithStdout(logger)
		assert.ErrorContains(t, action.Execute(), kase.errmsg)
		assert.Contains(t, logger.Outputs(), kase.output)
	}
}

func Test_Action_Can_Accept_StandardInput(t *testing.T) {
	logger := NewMockLogger()
	b := bytes.NewBufferString("hello\n")
	// we need to disable env expansion in statement so that `$s` is not replaced
	// with empty space (`$s` will be set during command execution, not before)
	action := NewAction("bash -c \"read s && echo $s\"", Env{}).WithStdin(b).WithEnvExpansion(false).WithStdout(logger)
	assert.NoError(t, action.Execute())
	assert.Contains(t, logger.Outputs(), "hello\n")
}
