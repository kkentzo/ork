package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Global struct {
	Default string            `yaml:"default"`
	Env     map[string]string `yaml:"env"`
}

type Orkfile struct {
	Global *Global `yaml:"global"`
	Tasks  []*Task `yaml:"tasks"`
}

func ParseOrkfile(path string) (*Orkfile, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	orkfile := &Orkfile{}
	if err := yaml.Unmarshal(contents, orkfile); err != nil {
		return nil, err
	}
	return orkfile, nil
}
