package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/apsdehal/go-logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Orkfiles(t *testing.T) {
	kases := []struct {
		test    string
		yml     string
		task    string
		outputs []string
	}{
		// ===================================
		{
			test: "env variables of dependencies are available within the task",
			yml: `
tasks:
  - name: foo
    depends_on:
      - bar
    actions:
      - echo $GLOBAL_ENV
  - name: bar
    env:
      - GLOBAL_ENV: foo
`,
			task:    "foo",
			outputs: []string{"foo\n"},
		},
		// ===================================
		{
			test: "local env overrides dependency env",
			yml: `
tasks:
  - name: foo
    depends_on:
      - bar
    env:
      - MY_VAR_1: foo
    actions:
      - echo ${MY_VAR_1}
  - name: bar
    env:
      - MY_VAR_1: bar
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
      - TASKS: $[echo version] $[echo version]
    actions:
      - echo ${TASKS}
`,
			task:    "foo",
			outputs: []string{"version version"},
		},
		// ===================================
		{
			test: "multiple command substitution in env",
			yml: `
tasks:
  - name: foo
    env:
      - FOO: $[echo foo]-$[echo bar]
    actions:
      - echo $FOO
`,
			task:    "foo",
			outputs: []string{"foo-bar"},
		},
		// ===================================
		{
			test: "hooks: run the proper hook set on success",
			yml: `
tasks:
  - name: foo
    actions:
      - echo foo
    on_success:
      - echo success
    on_failure:
      - echo failure
`,
			task:    "foo",
			outputs: []string{"foo", "success"},
		},
		// ===================================
		{
			test: "parent tasks' envs are visible in nested tasks",
			yml: `
tasks:
  - name: a
    env:
      - A_MY_VAR: a
    tasks:
      - name: b
        env:
          - B_MY_VAR: b
        tasks:
          - name: c
            env:
              - C_MY_VAR: c
            actions:
              - echo "${A_MY_VAR}${B_MY_VAR}${C_MY_VAR}"
`,
			task:    fmt.Sprintf("a%sb%sc", DEFAULT_TASK_GROUP_SEP, DEFAULT_TASK_GROUP_SEP),
			outputs: []string{"abc"},
		},
		// ===================================
		{
			test: "nested task env overrides the parent's env",
			yml: `
tasks:
  - name: foo
    env:
      - MY_VAR_3: foo
    actions:
      - echo $MY_VAR_3
    tasks:
      - name: bar
        env:
          - MY_VAR_3: bar
        actions:
          - echo $MY_VAR_3
        on_success:
          - echo success
        on_failure:
          - echo failure
`,
			task:    fmt.Sprintf("foo%sbar", DEFAULT_TASK_GROUP_SEP),
			outputs: []string{"foo", "bar", "success"},
		},
		// ===================================
		{
			test: "outer env variables are available in inner tasks regardless of lex order",
			yml: `
tasks:
  - name: foo
    depends_on:
      - bar
    env:
      - MY_VAR_4: ${MY_VAR_5}
    actions:
      - echo $MY_VAR_4
  - name: bar
    env:
      - MY_VAR_5: bar
`,
			task:    "foo",
			outputs: []string{"bar"},
		},

		// ===================================
		{
			test: "env expansion can be disabled",
			yml: `
tasks:
  - name: foo
    expand_env: false
    actions:
      - bash -c "for f in $(ls -1 main.go); do echo $f; done;"
`,
			task:    "foo",
			outputs: []string{"main.go"},
		},
		// ===================================
		{
			test: "env groups can see variables from the previous group",
			yml: `
global:
tasks:
  - name: foo
    env:
      - A: a
      - B: $[bash -c "echo $A"]
    actions:
      - echo $B
`,
			task:    "foo",
			outputs: []string{"a"},
		},
		// ===================================
		{
			test: "env can execute non-trivial bash statements",
			yml: `
tasks:
  - name: foo
    depends_on:
      - bar
    env_subst_greedy: true
    env:
      - MY_VAR_7: $[bash -c "if [ \"${MY_VAR_6}\" == \"production\" ]; then echo production; else echo staging; fi"]
    actions:
      - echo $MY_VAR_7
  - name: bar
    env:
      - MY_VAR_6: production
`,
			task:    "foo",
			outputs: []string{"production"},
		},
		// ===================================
		{
			test: "shell is parameterizable",
			yml: `
tasks:
  - name: foo
    actions:
      - python -c "import sys; sys.stdout.write('hello from python');"
`,
			task:    "foo",
			outputs: []string{"hello from python"},
		},
		// ===================================
		{
			test: "task dependency should have access to its env",
			yml: `
tasks:
  - name: parent
    env:
      - VAR: a
    tasks:
      - name: a
        actions:
          - echo "var=${VAR}"
  - name: child
    depends_on:
      - parent.a
`,
			task:    "child",
			outputs: []string{"var=a"},
		},
		// ===================================
		{
			test: "task env should expand its own env",
			yml: `
tasks:
  - name: yth
    env:
      - YTH: foo
    tasks:
      - name: cvs
        env:
          - CVS: ${YTH}
        actions:
          - echo "${CVS}"
`,
			task:    "yth.cvs",
			outputs: []string{"foo"},
		},
		// ===================================
		{
			test: "task names can contain the default separator",
			yml: fmt.Sprintf(`
tasks:
  - name: a%sb
    actions:
      - echo foo
`, DEFAULT_TASK_GROUP_SEP),
			task:    fmt.Sprintf("a%sb", DEFAULT_TASK_GROUP_SEP),
			outputs: []string{"foo"},
		},
		// ===================================
		{
			test: "task should run if a required env variable is available",
			yml: `
tasks:
  - name: kqs
    env:
      - TYUI: a
  - name: jho
    depends_on:
      - kqs
    require:
      exists:
        - TYUI
    actions:
      - echo $TYUI
`,
			task:    "jho",
			outputs: []string{"a\n"},
		},
		// ===================================
		{
			test: "task should run if a required env has the expected value",
			yml: `
tasks:
  - name: kkl
    env:
      - A: a
  - name: fgy
    depends_on:
      - kkl
    require:
      equals:
        A: a
    actions:
      - echo $A
`,
			task:    "fgy",
			outputs: []string{"a\n"},
		},
		// ===================================
		{
			test: "task should run if a required env has the expected calculated value",
			yml: `
tasks:
  - name: ght
    env:
      - QIO: a
        BVF: a
  - name: lch
    depends_on:
      - ght
    require:
      equals:
        BVF: ${QIO}
    actions:
      - echo $BVF
`,
			task:    "lch",
			outputs: []string{"a\n"},
		},
		// ===================================
		{
			test: "task with dot separator in its name should be able to identify its parent task",
			yml: `
tasks:
  - name: eyrwey
    generate:
      - name: sxbz
        env:
          - UYTRR: foo
      - name: dsldp
        env:
          - UYTRR: bar
    tasks:
      - name: wytedbzm.gsjaxa
        actions:
          - echo ${UYTRR}
`,
			task:    "eyrwey.sxbz.wytedbzm.gsjaxa",
			outputs: []string{"foo\n"},
		},
	}

	// set this to a kase.test value to run one test only
	only_kase := ""
	// set this to LOG_LEVEL_DEBUG to enable debug-level logging
	log_level := LOG_LEVEL_INFO

	for _, kase := range kases {
		if only_kase != "" && only_kase != kase.test {
			continue
		}
		log := NewMockLogger()
		log.SetLogLevel(log_level)
		// parse orkfile
		f := New()
		require.NoError(t, f.Parse([]byte(kase.yml)), kase.test)
		// execute task
		assert.NoError(t, f.RunTask(context.Background(), kase.task, log), kase.test)

		for _, ln := range log.Logs(logger.DebugLevel) {
			fmt.Println(ln)
		}
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
	assert.ErrorContains(t, f.RunTask(context.Background(), "foo", log), "cyclic dependency")
}

func Test_Orkfile_TaskShouldFail_WhenExistsRequirement_NotPresent(t *testing.T) {
	yml := `
tasks:
  - name: a
    require:
      exists:
        - FOO_1234
    actions:
      - echo $A
`
	f := New()
	assert.NoError(t, f.Parse([]byte(yml)))
	log := NewMockLogger()
	assert.ErrorContains(t, f.RunTask(context.Background(), "a", log), "failed requirement")
}

func Test_Orkfile_TaskShouldFail_WhenEqualsRequirement_NotPresent(t *testing.T) {
	yml := `
tasks:
  - name: azfw
    require:
      equals:
        GHTYT: kjaldasdashasjk
    actions:
      - echo ${GHTYT}
`
	f := New()
	assert.NoError(t, f.Parse([]byte(yml)))
	log := NewMockLogger()
	assert.ErrorContains(t, f.RunTask(context.Background(), "azfw", log), "expected value but does not exist")
}

func Test_Orkfile_TaskShouldFail_WhenEqualsRequirement_PresentButNotEqual(t *testing.T) {
	yml := `
tasks:
  - name: qoc
    env:
      - QOC: 5
  - name: sdw
    depends_on:
      - qoc
    require:
      equals:
        QOC: 6
    actions:
      - echo $QOC
`
	f := New()
	assert.NoError(t, f.Parse([]byte(yml)))
	log := NewMockLogger()
	assert.ErrorContains(t, f.RunTask(context.Background(), "sdw", log), "does not match its expected value")
}

func Test_Orkfile_Dependency_DoesNotExist(t *testing.T) {
	yml := `
tasks:
  - name: foo
    depends_on:
      - bar
`

	f := New()
	assert.NoError(t, f.Parse([]byte(yml)))
	log := NewMockLogger()
	assert.ErrorContains(t, f.RunTask(context.Background(), "foo", log), "dependency bar does not exist")
}

func Test_TaskAction_CanBeCancelled(t *testing.T) {
	yml := `
tasks:
  - name: read
    expand_env: false
    actions:
      - bash -c "while read s; do echo ${s}; done;"
`
	b := bytes.NewBufferString("")
	f := New().WithStdin(b)
	assert.NoError(t, f.Parse([]byte(yml)))

	log := NewMockLogger()
	ctx, cancel := context.WithCancel(context.Background())

	go func(o *Orkfile) {
		o.RunTask(ctx, "read", log)
	}(f)

	var err error
	_, err = b.WriteString("hello\n")
	assert.NoError(t, err)
	_, err = b.WriteString("goodbye\n")
	assert.NoError(t, err)

	// wait for the input to be ingested by the process
	time.Sleep(100 * time.Millisecond)

	// ok, let's cancel the process
	cancel()
	time.Sleep(50 * time.Millisecond)

	_, err = b.WriteString("this will not be in the output\n")
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	outputs := strings.Join(log.Outputs(), "")
	assert.Contains(t, outputs, "hello")
	assert.Contains(t, outputs, "goodbye")
	assert.NotContains(t, outputs, "this will not be in the output")
}

func Test_Orkfile_Task_Info(t *testing.T) {
	yml := `
tasks:
  - name: foo
    description: I am foo
`

	f := New()
	assert.NoError(t, f.Parse([]byte(yml)))
	assert.Equal(t, "[foo] I am foo", f.Info("foo"))
}

func Test_Orkfile_Task_Info_When_Task_DoesNot_Exist(t *testing.T) {
	f := New()
	assert.NoError(t, f.Parse([]byte("")))
	assert.Empty(t, f.Info("foo"))
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

func Test_Orkfile_Parse_On_Malformed_YML(t *testing.T) {
	invalid_yml := "invalid yaml contents"
	assert.Error(t, New().Parse([]byte(invalid_yml)))
}

func Test_Orkfile_Supports_Sequential_Env_Groups(t *testing.T) {
	env_items := ""
	template_val := ""
	target_val := ""
	for i := 0; i <= 20; i++ {
		env_items += fmt.Sprintf("      - VAR_%.2d: %d\n", i, i)
		template_val += fmt.Sprintf("$VAR_%.2d", i)
		target_val += fmt.Sprint(i)
	}

	yml := fmt.Sprintf(`
tasks:
  - name: env_ordering
    env:
%s
      - W_VAR: $[bash -c "echo %s"]
    actions:
      - echo $W_VAR
`, env_items, template_val)

	f := New()
	assert.NoError(t, f.Parse([]byte(yml)))
	log := NewMockLogger()

	assert.NoError(t, f.RunTask(context.Background(), "env_ordering", log))
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

	assert.NoError(t, f.RunTask(context.Background(), "dir", log))
	require.Equal(t, 1, len(log.Outputs()))
	assert.Contains(t, log.Outputs()[0], "hello")
}

func Test_Orkfile_Task_Failure_Hook_RunsOnError_And_Sets_ORK_ERROR(t *testing.T) {
	yml := `
tasks:
  - name: foo
    actions:
      - a_non_existent_program
    on_success:
      - echo success
    on_failure:
      - echo failure
      - echo $ORK_ERROR
`
	f := New()
	assert.NoError(t, f.Parse([]byte(yml)))
	log := NewMockLogger()

	assert.Error(t, f.RunTask(context.Background(), "foo", log))

	require.Equal(t, 2, len(log.Outputs()))
	assert.Contains(t, log.Outputs()[0], "failure")
	assert.Contains(t, log.Outputs()[1], "[foo] failed to start action: exec: a_non_existent_program: executable file not found")
}

func Test_Orkfile_Task_Does_Not_Exist(t *testing.T) {
	f := New()
	assert.NoError(t, f.Parse([]byte("")))
	assert.ErrorContains(t, f.RunTask(context.Background(), "foo", nil), "does not exist")
}

func Test_Orkfile_RunDefaultTask(t *testing.T) {
	yml := `
default: foo
tasks:
  - name: foo
    actions:
      - echo foo
`

	f := New()
	assert.NoError(t, f.Parse([]byte(yml)))
	log := NewMockLogger()

	assert.NoError(t, f.RunDefault(context.Background(), log))
	require.Equal(t, 1, len(log.Outputs()))
	assert.Contains(t, log.Outputs()[0], "foo")
}

func Test_Orkfile_RunDefaultTask_When_Task_DoesNot_Exist(t *testing.T) {
	f := New()
	assert.NoError(t, f.Parse([]byte("")))

	assert.ErrorContains(t, f.RunDefault(context.Background(), nil), "default task")
}

func Test_Orkfile_ListAllTasks(t *testing.T) {
	yml := `
tasks:
  - name: foo
  - name: bar
  - name: baz
`

	f := New()
	assert.NoError(t, f.Parse([]byte(yml)))

	tasks := f.GetTasks(All)
	assert.Equal(t, 3, len(tasks))
}

func Test_Orkfile_ActionableTasks(t *testing.T) {
	yml := `
tasks:
  - name: a1
    actions:
      - echo a1
  - name: a2
    depends_on:
      - a1
  - name: a3
    on_success:
      - echo a3
`
	f := New()
	assert.NoError(t, f.Parse([]byte(yml)))

	all := f.GetTasks(All)
	sort.Slice(all, func(i, j int) bool {
		return all[i].label < all[j].label
	})

	assert.Equal(t, 3, len(all))
	assert.True(t, all[0].IsActionable(), all[0].label)
	assert.True(t, all[1].IsActionable(), all[1].label)
	assert.False(t, all[2].IsActionable(), all[2].label)
}

func Test_Read(t *testing.T) {
	contents, err := Read("Orkfile.yml")
	assert.NoError(t, err)
	assert.NoError(t, New().Parse(contents))
}

func Test_Read_When_File_DoesNot_Exist(t *testing.T) {
	_, err := Read("this_file_does_not_exist")
	assert.Error(t, err)
}

func Test_Task_Generation(t *testing.T) {
	yml := `
tasks:
  - name: deploy
    env:
      - ACTION: deploy
    generate:
      - name: production
        env:
          - SERVER_URL: i_am_production
        actions:
          - echo $SERVER_URL
        on_success:
          - echo "production hook"
      - name: staging
        env:
          - SERVER_URL: i_am_staging
        actions:
          - echo $SERVER_URL
        on_success:
          - echo "staging hook"
    actions:
      - echo "deploy!"
    tasks:
      - name: ping
        actions:
          - echo "${ACTION} => pinging ${SERVER_URL}"
`
	f := New()
	assert.NoError(t, f.Parse([]byte(yml)))
	log := NewMockLogger()

	// do we have the correct tasks?
	all := f.GetTasks(All)
	sort.Slice(all, func(i, j int) bool {
		return all[i].label < all[j].label
	})

	names := []string{"deploy", "deploy.production", "deploy.production.ping", "deploy.staging", "deploy.staging.ping"}
	require.Equal(t, len(names), len(all))
	for i := range names {
		assert.Equal(t, names[i], all[i].label)
	}

	// ok, run the two tasks
	assert.NoError(t, f.RunTask(context.Background(), "deploy.production.ping", log))
	assert.NoError(t, f.RunTask(context.Background(), "deploy.staging.ping", log))

	// test the command outputs?
	expected := []string{
		"deploy!\n",
		"i_am_production\n",
		"production hook\n",
		"deploy => pinging i_am_production\n",
		"deploy!\n",
		"i_am_staging\n",
		"staging hook\n",
		"deploy => pinging i_am_staging\n",
	}
	actual := log.Outputs()

	require.Equal(t, len(expected), len(actual))
	for i := range actual {
		assert.Equal(t, expected[i], actual[i], actual[i])
	}

	// test the info logs
	expected = []string{
		"[deploy]",
		"[deploy.production]",
		"[deploy.production.ping]",
		"[deploy]",
		"[deploy.staging]",
		"[deploy.staging.ping]",
	}
	actual = log.Logs(logger.InfoLevel)
	require.Equal(t, len(expected), len(actual))
	for i := range actual {
		assert.True(t, strings.HasPrefix(actual[i], expected[i]), actual[i])
	}
}
