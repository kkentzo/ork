package main

import (
	"testing"

	"github.com/apsdehal/go-logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Orkfile_NoGlobalSection(t *testing.T) {
	yml := `
tasks:
  - name: foo
    actions:
      - echo foo
`
	log := NewMockLogger()
	f := New()
	assert.NoError(t, f.Parse([]byte(yml), log))
	assert.NoError(t, f.Task("foo").Execute())
	assert.Contains(t, log.Logs(logger.InfoLevel), "[foo] echo foo")
}

func Test_EmptyOrkfile(t *testing.T) {
	yml := ``
	log := NewMockLogger()
	f := New()
	assert.NoError(t, f.Parse([]byte(yml), log))
}

func Test_DefaultTask(t *testing.T) {
	yml := `
global:
  default: foo
tasks:
  - name: foo
    actions:
      - echo foo

`
	log := NewMockLogger()
	f := New()
	assert.NoError(t, f.Parse([]byte(yml), log))
	assert.NoError(t, f.DefaultTask().Execute())
	assert.Contains(t, log.Logs(logger.InfoLevel), "[foo] echo foo")
}

func Test_Orkfile_PrintsCommandOutput(t *testing.T) {
	yml := `
global:
  default: foo
tasks:
  - name: foo
    actions:
      - echo foo

`
	log := NewMockLogger()
	f := New()
	assert.NoError(t, f.Parse([]byte(yml), log))
	assert.NoError(t, f.Task("foo").Execute())
	assert.Contains(t, log.Logs(logger.InfoLevel), "[foo] echo foo")
	assert.Contains(t, log.Outputs(), "foo\n")
}

func Test_Orkfile_PreventsCyclicDependencyDetection(t *testing.T) {
	yml := `
tasks:
  - name: foo
    depends_on:
      - bar
    actions:
      - echo foo

  - name: bar
    depends_on:
      - foo
    actions:
      - echo bar
`
	log := NewMockLogger()
	f := New()
	assert.NoError(t, f.Parse([]byte(yml), log))
	err := f.Task("foo").Execute()
	assert.ErrorContains(t, err, "cyclic dependency")
}

func Test_Orkfile_UsesGlobalEnv(t *testing.T) {
	yml := `
global:
  env:
    GLOBAL_ENV: foo
tasks:
  - name: foo
    actions:
      - echo ${GLOBAL_ENV}
`
	log := NewMockLogger()
	f := New()
	assert.NoError(t, f.Parse([]byte(yml), log))
	assert.NoError(t, f.Task("foo").Execute())
	assert.Contains(t, log.Outputs(), "foo\n")
}

func Test_Orkfile_UsesLocalEnv(t *testing.T) {
	yml := `
tasks:
  - name: foo
    env:
      LOCAL_ENV: foo
    actions:
      - echo $LOCAL_ENV
`
	log := NewMockLogger()
	f := New()
	assert.NoError(t, f.Parse([]byte(yml), log))
	assert.NoError(t, f.Task("foo").Execute())
	assert.Contains(t, log.Outputs(), "foo\n")
}

func Test_Orkfile_LocalEnv_Overrides_GlobalEnv(t *testing.T) {
	yml := `
global:
  env:
    VARIABLE: foo
tasks:
  - name: foo
    env:
      VARIABLE: bar
    actions:
      - echo ${VARIABLE}
`
	log := NewMockLogger()
	f := New()
	assert.NoError(t, f.Parse([]byte(yml), log))
	assert.NoError(t, f.Task("foo").Execute())
	assert.Contains(t, log.Outputs(), "bar\n")
}

func Test_Orkfile_Env_Does_CommandSubstitution(t *testing.T) {
	yml := `
tasks:
  - name: foo
    env:
      BAR: $(echo bar)
    actions:
      - 'bash -c "echo $BAR"'
`
	log := NewMockLogger()
	f := New()
	assert.NoError(t, f.Parse([]byte(yml), log))
	assert.NoError(t, f.Task("foo").Execute())
	assert.Contains(t, log.Outputs(), "bar\n")
	assert.Contains(t, log.Logs(logger.InfoLevel), "[foo] echo $BAR")
}

func Test_Orkfile_Parse_Fails_When_TwoTasks_Exist_WithTheSameName(t *testing.T) {
	yml := `
tasks:
  - name: foo
    actions:
      - echo foo1
  - name: foo
    actions:
      - echo foo2
`
	log := NewMockLogger()
	f := New()
	assert.ErrorContains(t, f.Parse([]byte(yml), log), "duplicate task")

}

func Test_Orkfile_TaskNotFound(t *testing.T) {
	yml := ``
	log := NewMockLogger()
	f := New()
	assert.NoError(t, f.Parse([]byte(yml), log))
	assert.Nil(t, f.Task("foo"))
}

func Test_Orkfile_Support_ArbitraryShell(t *testing.T) {
	yml := `
tasks:
  - name: run_project_orkfile
    actions:
      - go run .
`
	log := NewMockLogger()
	f := New()
	assert.NoError(t, f.Parse([]byte(yml), log))
	assert.NoError(t, f.Task("run_project_orkfile").Execute())
	require.Equal(t, 1, len(log.Outputs()))
	assert.Contains(t, log.Outputs()[0], "[build] go build")
	assert.Contains(t, log.Logs(logger.InfoLevel), "[run_project_orkfile] go run .")
}
