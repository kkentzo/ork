package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_envValues(t *testing.T) {
	kases := []struct {
		statement string
		greedy    bool
		expected  []envToken
	}{
		{"", false, []envToken{{"", false}}},
		{"12 12", false, []envToken{{"12 12", false}}},
		{"$[echo foo]", false, []envToken{{"echo foo", true}}},
		{`$[bash -c "echo $(echo foo)"]`, false, []envToken{{"bash -c \"echo $(echo foo)\"", true}}},
		{"1-$[foo]-2-$[echo foo ]-3-$[bar]-4", false, []envToken{
			{"1-", false},
			{"foo", true},
			{"-2-", false},
			{"echo foo ", true},
			{"-3-", false},
			{"bar", true},
			{"-4", false},
		}},
		{`$[bash -c "if [ foo = foo]; then echo foo; else echo bar; fi"]`, true, []envToken{
			{"bash -c \"if [ foo = foo]; then echo foo; else echo bar; fi\"", true},
		}},
	}

	for idx, kase := range kases {
		assert.Equal(t, kase.expected, parseEnvTokens(kase.statement, kase.greedy), fmt.Sprintf("case index=%d", idx))
	}
}
