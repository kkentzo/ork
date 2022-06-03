package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Shell_Command_Logs(t *testing.T) {
	log := NewMockLogger()
	assert.NoError(t, Execute("/bin/bash", "echo foo", log))
	assert.Contains(t, log.Outputs(), "foo\n")
}

func Test_Shell_InheritsEnvironment(t *testing.T) {
	os.Setenv("A_RANDOM_VAR", "$(echo foo)")
	log := NewMockLogger()
	assert.NoError(t, Execute("/bin/bash", "echo ${A_RANDOM_VAR}", log))
	assert.Contains(t, log.Outputs(), "foo\n")
}

func Test_Shell_Performs_CommandSubstitution(t *testing.T) {
	log := NewMockLogger()
	assert.NoError(t, Execute("/bin/bash", "echo $(echo foo)", log))
	assert.Contains(t, log.Outputs(), "foo\n")
}
