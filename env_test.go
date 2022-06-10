package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_envValues(t *testing.T) {
	kases := []struct {
		statement string
		expected  []envToken
	}{
		{"", []envToken{{"", false}}},
		{"12 12", []envToken{{"12 12", false}}},
		{"$[echo foo]", []envToken{{"echo foo", true}}},
		{"$[bash -c \"echo $(echo foo)\"]", []envToken{{"bash -c \"echo $(echo foo)\"", true}}},
		{"1-$[foo]-2-$[echo foo ]-3-$[bar]-4", []envToken{
			{"1-", false},
			{"foo", true},
			{"-2-", false},
			{"echo foo ", true},
			{"-3-", false},
			{"bar", true},
			{"-4", false},
		}},
	}

	for idx, kase := range kases {
		assert.Equal(t, kase.expected, parseEnvTokens(kase.statement), fmt.Sprintf("case index=%d", idx))
	}
}
