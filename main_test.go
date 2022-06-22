package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Ork_Command(t *testing.T) {
	kases := []struct {
		description string
		args        []string // do not include the executable
		output      []string
	}{
		{"info for single task",
			[]string{"-i", "build"},
			[]string{"[build] build the application\n"},
		},
	}
	for _, kase := range kases {
		logger := NewMockLogger()
		kase.args = append([]string{"exe"}, kase.args...)
		assert.NoError(t, runApp(kase.args, logger))
		out := logger.Outputs()
		assert.Equal(t, len(kase.output), len(out))
		for i := 0; i < len(kase.output); i++ {
			assert.Equal(t, kase.output[i], out[i])
		}
	}
}
