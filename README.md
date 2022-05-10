[![test](https://github.com/kkentzo/ork/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/kkentzo/ork/actions/workflows/test.yml)

# ork

`ork` is a program to organize, compose and execute command workflows
in software projects.

It is meant as a light-weight substitute to the venerable `Makefile`
program and especially for `Makefile`s that make heavy use of `.PHONY`
targets (i.e. `Makefile`s that focus on operational tasks rather than
files).

`ork` is simple and mostly inspired by the workflow syntax of
[Github](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)
and [Gitlab](https://docs.gitlab.com/ee/ci/yaml/gitlab_ci_yaml.html)
CI products.

## `Orkfile`

`ork` is based on project-specific Orkfiles (`Orkfile.yml`) which
specify and set up the various tasks and their
dependencies. `Orkfile`s consist of two main sections:

- a `global` section where general settings are specified (shell, environment etc.)
- a `tasks` section which defines the various workflows that can be executed

Here's an example for a golang application:


```yaml
global:
  # the default task
  default: build
  # shell to execute commands into
  shell: /bin/bash
  # environment variables available to all tasks
  env:
    GO_BUILD: go build
    GO_TARGET: bin/ork

# task definitions
# each task can have its own environment variables and
# a bunch of actions to be executed in the globally defined shell (default: bash)
tasks:

  - name: build
    description: build the application
    env:
      GOOS: linux
      GOARCH: amd64
    actions:
      - ${GO_BUILD} -o ${GO_TARGET}

  - name: test
    description: test the application
    depends_on:
      - build
    actions:
      - go test . -v -count=1

  - name: clean
    actions:
      - rm -rf bin
```

## Installation & Usage

`ork` can be installed by downloading the latest release binary from
[here](https://github.com/kkentzo/ork/releases) and putting it in a
convenient place in your `$PATH` (e.g. `/usr/local/bin`).

Run `ork -h` for program options.

## Autocompletion

`ork` supports task autocompletion in the command-line. Follow the
guides below to enable system-wide auto-completion (courtesy of the
excellent [cli](https://github.com/urfave/cli) library):

- [bash instructions](https://github.com/urfave/cli/blob/main/docs/v2/manual.md#distribution-and-persistent-autocompletion)
- [zsh support](https://github.com/urfave/cli/blob/main/docs/v2/manual.md#zsh-support)
- [PowerShell support](https://github.com/urfave/cli/blob/main/docs/v2/manual.md#powershell-support)
