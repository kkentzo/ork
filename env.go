package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

type Env map[string]string

// create and return a copy of `this`
func (this Env) Copy() Env {
	new := Env{}
	// first set all of `this`
	for k, v := range this {
		new[k] = v
	}
	return new
}

// apply all the entries of `this` to the actual environment
// all env values will be parsed to detect substitution patterns $[...]
// which will be executed as actions whose output will be interpolated in the env value
func (this Env) Apply() error {
	env := []struct {
		key   string
		value string
	}{}

	// assemble env into an array
	for ek, ev := range this {
		env = append(env, struct {
			key   string
			value string
		}{ek, ev})
	}
	// sort array lexicographically on keys
	sort.Slice(env, func(i, j int) bool {
		return env[i].key < env[j].key
	})

	// apply sorted key, value entries
	for _, kv := range env {
		val := ""
		for _, token := range parseEnvTokens(kv.value) {
			v, err := token.expand()
			if err != nil {
				return fmt.Errorf("key %s: %s: %v", kv.key, kv.value, err)
			}

			val += v
		}
		os.Setenv(kv.key, val)
	}

	return nil
}

// add all the entries of `other` to `this` by overriding them if already present
func (this Env) Merge(other Env) Env {
	for k, v := range other {
		this[k] = v
	}
	return this
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
func parseEnvTokens(statement string) []envToken {
	re := regexp.MustCompile(`\$\[.*?\]`)

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
		action := NewAction(e.value).WithStdout(buf).WithEnvExpansion(false)
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
