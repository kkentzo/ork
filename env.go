package main

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

// add all the entries of `other` to `this` by overriding them if already present
func (this Env) Merge(other Env) {
	for k, v := range other {
		this[k] = v
	}
}
