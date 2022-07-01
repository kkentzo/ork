package main

import (
	"context"
	"os"
	"testing"

	"github.com/apsdehal/go-logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var yml string = `
tasks:
  - name: foo
    description: i am foo
    env:
      - VAR: foo
    actions:
      - echo $VAR
    tasks:
      - name: bar
        env:
          - VAR: bar
        actions:
          - echo $VAR
`

func Test_Ork_Command(t *testing.T) {
	orkfile_path := "Orkfile.command_tests.yml"
	os.WriteFile(orkfile_path, []byte(yml), os.ModePerm)
	defer os.Remove(orkfile_path)

	kases := []struct {
		description string
		args        []string // do not include the executable
		output      []string
	}{
		{
			"info for single task",
			[]string{"-i", "foo"},
			[]string{"[foo] i am foo\n"},
		},
		{
			"list all tasks in lexicographic order",
			[]string{"-i"},
			[]string{
				"[foo] i am foo\n",
				"[foo.bar] <no description>\n",
			},
		},
		{
			"execute single task",
			[]string{"foo"},
			[]string{"foo\n"},
		},
		{
			"execute multiple tasks",
			[]string{"foo", "foo.bar"},
			[]string{"foo\n", "foo\n", "bar\n"},
		},
	}
	for _, kase := range kases {
		logger := NewMockLogger()
		kase.args = append([]string{"exe", "-p", orkfile_path}, kase.args...)
		require.NoError(t, runApp(context.Background(), kase.args, logger), kase.description)
		out := logger.Outputs()
		require.Equal(t, len(kase.output), len(out), kase.description)
		for i := 0; i < len(kase.output); i++ {
			assert.Equal(t, kase.output[i], out[i], kase.description)
		}
	}
}

func Test_Ork_Command_Errors(t *testing.T) {
	orkfile_path := "Orkfile.command_errors.yml"
	os.WriteFile(orkfile_path, []byte(yml), os.ModePerm)
	defer os.Remove(orkfile_path)

	kases := []struct {
		description string
		args        []string // do not include the executable
		errmsg      string
	}{
		{
			"requested task does not exist",
			[]string{"does_not_exist"},
			"task does_not_exist does not exist",
		},
		{
			"default task does not exist",
			[]string{},
			"default task has not been set",
		},
	}
	for _, kase := range kases {
		log := NewMockLogger()
		kase.args = append([]string{"exe", "-p", orkfile_path}, kase.args...)
		err := runApp(context.Background(), kase.args, log)
		require.Error(t, err, kase.description)
		assert.Equal(t, kase.errmsg, err.Error(), kase.description)
	}
}

func Test_Ork_Command_MalformedOrkfile(t *testing.T) {
	orkfile_path := "Orkfile.malformed_json.yml"
	os.WriteFile(orkfile_path, []byte("invalid_yaml_contents"), os.ModePerm)
	defer os.Remove(orkfile_path)

	log := NewMockLogger()
	args := []string{"exe", "-p", orkfile_path}
	err := runApp(context.Background(), args, log)
	assert.ErrorContains(t, err, "failed to parse Orkfile")
}

func Test_Ork_Command_LogLevel(t *testing.T) {
	orkfile_path := "Orkfile.command_log_level.yml"
	os.WriteFile(orkfile_path, []byte(yml), os.ModePerm)
	defer os.Remove(orkfile_path)

	log := NewMockLogger()

	// let's try an invalid log level
	args := []string{"exe", "-p", orkfile_path, "-l", "invalid"}
	err := runApp(context.Background(), args, log)
	assert.ErrorContains(t, err, "unknown log level: invalid")

	// let's try the default log level
	args = []string{"exe", "-p", orkfile_path, "foo"}
	assert.NoError(t, runApp(context.Background(), args, log))
	assert.Empty(t, log.Logs(logger.DebugLevel))

	// let's try the debug log level
	args = []string{"exe", "-p", orkfile_path, "-l", "debug", "foo"}
	assert.NoError(t, runApp(context.Background(), args, log))
	assert.NotEmpty(t, log.Logs(logger.DebugLevel))
}

func Test_Ork_Command_Search(t *testing.T) {
	orkfile_path := "Orkfile.search_tests.yml"
	os.WriteFile(orkfile_path, []byte(yml), os.ModePerm)
	defer os.Remove(orkfile_path)

	kases := []struct {
		description string
		term        string
		results     []string
	}{
		{"contains foo", "foo", []string{"[foo] i am foo\n", "[foo.bar] <no description>\n"}},
		{"match foo and bar", "foo(\\.bar)?", []string{"[foo] i am foo\n", "[foo.bar] <no description>\n"}},
		{"foo only", "^foo$", []string{"[foo] i am foo\n"}},
		{"match bar but not foo", "bar", []string{"[foo.bar] <no description>\n"}},
		{"no match", "baz", []string{}},
	}

	for _, kase := range kases {
		logger := NewMockLogger()
		args := []string{"exe", "-p", orkfile_path, "-s", kase.term}
		require.NoError(t, runApp(context.Background(), args, logger), kase.description)

		out := logger.Outputs()
		require.Equal(t, len(kase.results), len(out), kase.description)
		for i := 0; i < len(kase.results); i++ {
			assert.Equal(t, kase.results[i], out[i], kase.description)
		}

	}
}

func Test_Ork_Command_Search_Error(t *testing.T) {
	orkfile_path := "Orkfile.search_tests_error.yml"
	os.WriteFile(orkfile_path, []byte(yml), os.ModePerm)
	defer os.Remove(orkfile_path)

	kases := []struct {
		description string
		term        string
		errmsg      string
	}{
		{"invalid regex", `g(-z]+ng`, "invalid regular expression"},
		{"no search term provided", "", "no search term provided"},
	}

	for _, kase := range kases {
		logger := NewMockLogger()
		args := []string{"exe", "-p", orkfile_path, "-s", kase.term}
		err := runApp(context.Background(), args, logger)

		assert.ErrorContains(t, err, kase.errmsg)
	}
}
