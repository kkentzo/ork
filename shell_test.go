package main

import (
	"fmt"
	"io"
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
		assert.NoError(t, Execute(kase.shell, kase.statement, kase.env, logger, nil), fmt.Sprintf("test case: %d", idx))
		assert.Contains(t, logger.Outputs(), kase.output, fmt.Sprintf("test case: %d", idx))
	}
}

func Test_Bash_Shell_Execute_Errors(t *testing.T) {
	kases := []struct {
		shell     string
		statement string
		env       map[string]string
		stdin     io.Reader
		errmsg    string
		output    string
	}{
		// command prints error
		{"/bin/bash", "echo foo; exit 1", map[string]string{}, nil, "exit status 1", "foo\n"},
	}

	for _, kase := range kases {
		logger := NewMockLogger()
		err := Execute(kase.shell, kase.statement, kase.env, logger, kase.stdin)
		assert.ErrorContains(t, err, kase.errmsg)
		assert.Contains(t, logger.Outputs(), kase.output)
	}
}
