package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// `Env` maps are **not** to be mutated once constructed
type Env map[string]string

// apply all the entries of `this` to the actual environment
// does not mutate `this` in any way
// all env values will be parsed to detect substitution patterns $[...]
// which will be executed as actions whose output will be interpolated in the env value
func (this Env) Apply(greedyEnvSubst bool) error {
	// apply key, value entries
	for key, value := range this {
		val := ""
		for _, token := range parseEnvTokens(value, greedyEnvSubst) {
			v, err := token.expand()
			if err != nil {
				return fmt.Errorf("key %s: %s: %v", key, value, err)
			}

			val += v
		}
		os.Setenv(key, val)
	}

	return nil
}

// this represents a portion of an environment variable's value
// that will either be executed and replaced with the execution output
// or will just be used as is
type envToken struct {
	value    string
	isAction bool
}

// split the statement into discrete tokens that will:
// - either be executed and replaced with the execution output
// - or will just be used as is
func parseEnvTokens(statement string, greedyEnvSubst bool) []envToken {
	var re *regexp.Regexp
	if greedyEnvSubst {
		re = regexp.MustCompile(`\$\[.*\]+`)
	} else {
		re = regexp.MustCompile(`\$\[.*?\]`)
	}

	tokens := []envToken{}
	matches := re.FindAllStringIndex(statement, -1)
	if matches == nil {
		return []envToken{{statement, false}}
	}
	c := 0
	r1, r2 := 2, 1
	for _, loc := range matches {
		if loc[0] > c {
			tokens = append(tokens, envToken{statement[c:loc[0]], false})
		}
		c = loc[1]
		tokens = append(tokens, envToken{statement[loc[0]+r1 : loc[1]-r2], true})
	}
	if c < len(statement) {
		tokens = append(tokens, envToken{statement[c:len(statement)], false})
	}
	return tokens
}

// return the token's representation
// either by executing the command
// or by returning its value as is
func (e envToken) expand() (out string, err error) {
	if e.isAction {
		buf := bytes.NewBuffer([]byte{})
		action := NewAction(context.Background(), e.value).WithStdout(buf).WithEnvExpansion(false)
		if err = action.Execute(); err != nil {
			return
		}
		out = buf.String()

	} else {
		out = e.value
	}

	if strings.HasSuffix(out, "\n") {
		out = out[:len(out)-1]
	}

	return
}
