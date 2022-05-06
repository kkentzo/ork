package main

import (
	"testing"

	"github.com/apsdehal/go-logger"
	"github.com/stretchr/testify/assert"
)

func Test_Orkfile_NoGlobalSection(t *testing.T) {
	yml := `
tasks:
  - name: foo
    actions:
      - echo foo
`
	log := NewMockLogger()
	f := New(log)
	assert.NoError(t, f.Parse([]byte(yml)))
	assert.NoError(t, f.Task("foo").Execute())
	assert.Contains(t, log.Logs(logger.InfoLevel), "[foo] echo foo")
}

func Test_EmptyOrkfile(t *testing.T) {
	yml := ``
	log := NewMockLogger()
	f := New(log)
	assert.NoError(t, f.Parse([]byte(yml)))
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
	f := New(log)
	assert.NoError(t, f.Parse([]byte(yml)))
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
	f := New(log)
	assert.NoError(t, f.Parse([]byte(yml)))
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
	f := New(log)
	assert.NoError(t, f.Parse([]byte(yml)))
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
	f := New(log)
	assert.NoError(t, f.Parse([]byte(yml)))
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
      - echo ${LOCAL_ENV}
`
	log := NewMockLogger()
	f := New(log)
	assert.NoError(t, f.Parse([]byte(yml)))
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
	f := New(log)
	assert.NoError(t, f.Parse([]byte(yml)))
	assert.NoError(t, f.Task("foo").Execute())
	assert.Contains(t, log.Outputs(), "bar\n")

}
