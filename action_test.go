package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Action_Execute_Errors(t *testing.T) {
	noOutput := ""
	kases := []struct {
		statement string
		errmsg    string
		output    string
	}{
		// command prints error
		{"bash -c \"echo foo; exit 1\"", "exit status 1", "foo\n"},
	}

	for _, kase := range kases {
		logger := NewMockLogger()
		action := NewAction(context.Background(), kase.statement).WithStdout(logger)
		assert.ErrorContains(t, action.Execute(), kase.errmsg)
		if kase.output != noOutput {
			assert.Contains(t, logger.Outputs(), kase.output)
		}
	}
}

func Test_Action_Can_Accept_StandardInput(t *testing.T) {
	logger := NewMockLogger()
	b := bytes.NewBufferString("hello\n")
	// we need to disable env expansion in statement so that `$s` is not replaced
	// with empty space (`$s` will be set during command execution, not before)
	action := NewAction(context.Background(), "bash -c \"read s && echo $s\"").
		WithStdin(b).
		WithEnvExpansion(false).
		WithStdout(logger)
	assert.NoError(t, action.Execute())
	assert.Contains(t, logger.Outputs(), "hello\n")
}
