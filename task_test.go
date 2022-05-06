package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_execute_OnSuccess(t *testing.T) {
	statement := "${DOES_NOT_EXIST} foo"
	out, err := execute("bash", statement, []string{"DOES_NOT_EXIST=echo"})
	assert.Equal(t, "foo\n", out)
	assert.NoError(t, err)
}

func Test_execute_OnError(t *testing.T) {
	statement := "${ENV_DOES_NOT_EXIST} foo"
	out, err := execute("bash", statement, []string{})
	assert.Empty(t, out)
	assert.Error(t, err)
	assert.Equal(t, "exit status 127", err.Error())
}

func Test_execute_UsesGlobalEnv(t *testing.T) {
	os.Setenv("GLOBAL_ENVIRONMENT_VARIABLE", "foo")
	statement := "echo \"global=${GLOBAL_ENVIRONMENT_VARIABLE}\""
	out, err := execute("bash", statement, []string{})
	assert.Equal(t, "global=foo\n", out)
	assert.NoError(t, err)
}

func Test_execute_UsesLocalEnv(t *testing.T) {
	statement := "echo \"local=${LOCAL_ENVIRONMENT_VARIABLE}\""
	out, err := execute("bash", statement, []string{"LOCAL_ENVIRONMENT_VARIABLE=bar"})
	assert.Equal(t, "local=bar\n", out)
	assert.NoError(t, err)
}

func Test_execute_LocalEnv_Overrides_GlobalEnv(t *testing.T) {
	os.Setenv("GLOBAL_ENVIRONMENT_VARIABLE", "foo")
	statement := "echo \"global=${GLOBAL_ENVIRONMENT_VARIABLE}\""
	out, err := execute("bash", statement, []string{"GLOBAL_ENVIRONMENT_VARIABLE=bar"})
	assert.Equal(t, "global=bar\n", out)
	assert.NoError(t, err)
}

func Test_Task_BuildSelf(t *testing.T) {
	cfg, err := ParseOrkfile("Orkfile.yml")
	assert.NoError(t, err)
	r := NewTaskRegistry(cfg)
	build := r["build"]
	assert.NoError(t, build.Execute())
}
