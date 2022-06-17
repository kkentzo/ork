package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Action_Execute(t *testing.T) {
	kases := []struct {
		statement string
		env       Env
		expandEnv bool
		output    string
	}{
		// action prints output
		{"echo foo\n", Env{}, true, "foo\n"},
		// action can see the environment
		{"echo $A_RANDOM_VAR", Env{"A_RANDOM_VAR": "foo"}, true, "foo\n"},
		// action can execute arbitrary commands
		{"python -c \"import sys; sys.stdout.write('hello from python');\"", Env{}, true, "hello from python"},
		// env does multiple command substitution
		{"echo $FOO", Env{"FOO": "$[echo foo]-$[echo bar]"}, true, "foo-bar\n"},
		// nested command substitutions are executed properly
		{"echo $FOO", Env{"FOO": "$[bash -c \"echo $(echo foo)\"]-$[echo bar]"}, true, "foo-bar\n"},
		// complex bash scripts are executed properly
		{`bash -c "if [ \"$A_VAR\" == \"i am foo\" ]; then echo $(echo \"yes, $A_VAR\"); else echo \"no, i am not foo\"; fi"`, Env{"A_VAR": "i am foo"}, true, "yes, i am foo\n"},
		// bash for loop
		{`bash -c "for f in $(ls -1 main.go); do echo $f; done;"`, Env{}, false, "main.go\n"},
	}

	for idx, kase := range kases {
		logger := NewMockLogger()
		assert.NoError(t, kase.env.Apply())
		action := NewAction(kase.statement).WithStdout(logger).WithEnvExpansion(kase.expandEnv)
		assert.NoError(t, action.Execute(), fmt.Sprintf("test case: %d", idx))
		assert.Contains(t, logger.Outputs(), kase.output, fmt.Sprintf("test case: %d", idx))
	}
}

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
		action := NewAction(kase.statement).WithStdout(logger)
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
	action := NewAction("bash -c \"read s && echo $s\"").WithStdin(b).WithEnvExpansion(false).WithStdout(logger)
	assert.NoError(t, action.Execute())
	assert.Contains(t, logger.Outputs(), "hello\n")
}
