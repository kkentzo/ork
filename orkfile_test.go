package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Orkfiles(t *testing.T) {
	kases := []struct {
		test    string
		yml     string
		task    string
		errmsg  string
		outputs []string
	}{
		// ===================================
		{
			test: "no global section",
			yml: `
tasks:
  - name: foo
    actions:
      - echo foo
`,
			task:    "foo",
			outputs: []string{"foo\n"},
		},
		// ===================================
		{
			test: "default task",
			yml: `
global:
  default: foo
tasks:
  - name: foo
    actions:
      - echo foo
`,
			task:    "foo",
			outputs: []string{"foo\n"},
		},
		// ===================================
		{
			test: "uses global env",
			yml: `
global:
  env:
    GLOBAL_ENV: foo
tasks:
  - name: foo
    actions:
      - echo ${GLOBAL_ENV}
`,
			task:    "foo",
			outputs: []string{"foo\n"},
		},
		// ===================================
		{
			test: "local env overrides global env",
			yml: `
global:
  env:
    MY_VAR: bar
tasks:
  - name: foo
    env:
      MY_VAR: foo
    actions:
      - echo ${MY_VAR}
`,
			task:    "foo",
			outputs: []string{"foo\n"},
		},
		// ===================================
		{
			test: "command substitution in env",
			yml: `
tasks:
  - name: foo
    env:
      TASK: $[echo clean] $[echo clean]
    actions:
      - go run . $TASK
`,
			task:    "foo",
			outputs: []string{"rm -rf bin", "rm -rf bin"},
		},
		// ===================================
	}

	for _, kase := range kases {
		log := NewMockLogger()
		// parse orkfile
		f := New()
		require.NoError(t, f.Parse([]byte(kase.yml)), kase.test)
		// find and execute task
		task := f.Task(kase.task)
		require.NotNil(t, task, kase.test)
		assert.NoError(t, task.Execute(f.Env(), log), kase.test)
		// check expected outputs
		outputs := log.Outputs()
		require.Equal(t, len(kase.outputs), len(outputs), kase.test)
		for idx := 0; idx < len(outputs); idx++ {
			assert.Contains(t, outputs[idx], kase.outputs[idx], kase.test)
		}
	}
}

func Test_EmptyOrkfile(t *testing.T) {
	assert.NoError(t, New().Parse([]byte("")))
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

	f := New()
	assert.NoError(t, f.Parse([]byte(yml)))
	log := NewMockLogger()
	assert.ErrorContains(t, f.Task("foo").Execute(f.Env(), log), "cyclic dependency")
}

func Test_Orkfile_GlobalEnv_OverridenByLocalEnv_PerTask(t *testing.T) {
	yml := `
global:
  env:
    FOO: foo
tasks:
  - name: bar
    env:
      FOO: bar
    actions:
      - echo $FOO
  - name: foo
    actions:
      - echo $FOO
`
	f := New()
	assert.NoError(t, f.Parse([]byte(yml)))
	log := NewMockLogger()

	assert.NoError(t, f.Task("bar").Execute(f.Env(), log))
	assert.NoError(t, f.Task("foo").Execute(f.Env(), log))

	outputs := log.Outputs()
	require.Equal(t, 2, len(outputs))
	assert.Equal(t, "bar\n", outputs[0])
	assert.Equal(t, "foo\n", outputs[1])
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

	f := New()
	assert.ErrorContains(t, f.Parse([]byte(yml)), "duplicate task")
}

func Test_Orkfile_Supports_Env_Ordering(t *testing.T) {
	env_items := ""
	template_val := ""
	target_val := ""
	for i := 0; i <= 20; i++ {
		env_items += fmt.Sprintf("      VAR_%.2d: %d\n", i, i)
		template_val += fmt.Sprintf("$VAR_%.2d", i)
		target_val += fmt.Sprint(i)
	}

	yml := fmt.Sprintf(`
tasks:
  - name: env_ordering
    env:
      W_VAR: %s
%s
    actions:
      - bash -c "echo $W_VAR"
`, template_val, env_items)

	f := New()
	assert.NoError(t, f.Parse([]byte(yml)))
	log := NewMockLogger()

	assert.NoError(t, f.Task("env_ordering").Execute(f.Env(), log))
	require.Equal(t, 1, len(log.Outputs()))
	assert.Contains(t, log.Outputs()[0], target_val)
}

func Test_Orkfile_Task_With_Working_Directory(t *testing.T) {
	os.Mkdir("test_foo", os.ModePerm)
	os.WriteFile("test_foo/bar", []byte("hello"), os.ModePerm)
	defer os.RemoveAll("test_foo")

	yml := `
tasks:
  - name: dir
    working_dir: test_foo
    actions:
      - cat bar
`
	f := New()
	assert.NoError(t, f.Parse([]byte(yml)))
	log := NewMockLogger()

	assert.NoError(t, f.Task("dir").Execute(f.Env(), log))
	require.Equal(t, 1, len(log.Outputs()))
	assert.Contains(t, log.Outputs()[0], "hello")
}
