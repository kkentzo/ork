package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Shell_Command_Logs(t *testing.T) {
	log := NewMockLogger()
	assert.NoError(t, Execute("/bin/bash", map[string]string{}, "echo foo", log))
	assert.Contains(t, log.Outputs(), "foo\n")
}

func Test_Shell_Env_Variable(t *testing.T) {
	log := NewMockLogger()
	assert.NoError(t, Execute("/bin/bash", map[string]string{"FOO": "foo", "BAR": "bar"}, "echo ${FOO} $BAR", log))
	assert.Contains(t, log.Outputs(), "foo bar\n")
}

func Test_Shell_InheritsEnvironment(t *testing.T) {
	os.Setenv("A_RANDOM_VAR", "$(echo foo)")
	log := NewMockLogger()
	assert.NoError(t, Execute("/bin/bash", map[string]string{}, "echo ${A_RANDOM_VAR}", log))
	assert.Contains(t, log.Outputs(), "foo\n")
}

func Test_Shell_Performs_CommandSubstitution(t *testing.T) {
	log := NewMockLogger()
	assert.NoError(t, Execute("/bin/bash", map[string]string{}, "echo $(echo foo)", log))
	assert.Contains(t, log.Outputs(), "foo\n")
}

func Test_Shell_Performs_CommandSubstitution_InEnv(t *testing.T) {
	log := NewMockLogger()
	assert.NoError(t, Execute("/bin/bash", map[string]string{"FOO": "$(echo foo)$(echo bar)"}, "echo $FOO", log))
	assert.Contains(t, log.Outputs(), "foobar\n")
}
