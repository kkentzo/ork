[![test](https://github.com/kkentzo/ork/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/kkentzo/ork/actions/workflows/test.yml)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/kkentzo/ork)

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

## Features

### Workflow organization

`ork` organizes workflows in yaml format (`Orkfile.yml`) as a
collection of tasks, each consisting of a sequence of actions that are
executed independently of each other (i.e. in separate processes). For
example:

```yaml
global:
  default: build
  env:
    GO_BUILD: go -ldflags="-s -w"

tasks:

  - name: build
    description: build the application
    env:
      GOOS: linux
      GOARCH: amd64
      GO_TARGET: bin/foo
    actions:
      - $GO_BUILD -o $GO_TARGET

```

A task can contain any kind of executable actions, e.g.:

```yaml
tasks:
  - name: say-hello
    description: Say hello from python
    actions:
      - python -c "import sys; sys.stdout.write('hello from python')"
```

### Global configuration

As seen above, Orkfiles can also have a global section for setting up
properties for all tasks (that, nevertheless, can still be overriden
by individual tasks) such as environment variables, the default task
etc.

### Task dependencies

`ork` also supports task dependencies (with cyclic dependency
detection), for example:

```yaml
tasks:

  - name: build
    description: build the application
    actions:
      - ...

  - name: test
    description: test the application
    actions:
      - ...

  - name: deploy
    description: deploy the application
    depends_on:
      - build
      - test
    actions:
      - ...
```

## Installation & Usage

`ork` can be installed by downloading the latest release binary from
[here](https://github.com/kkentzo/ork/releases) and putting it in a
convenient place in your `$PATH` (e.g. `/usr/local/bin`).

`ork` can execute one or more tasks defined in an Orkfile by running:

```bash
$ ork task1 task2 ...
```

Run `ork -h` for program options.

## Autocompletion

`ork` supports task autocompletion in the command-line. Follow the
guides below to enable system-wide auto-completion (courtesy of the
excellent [cli](https://github.com/urfave/cli) library):

- [bash instructions](https://github.com/urfave/cli/blob/main/docs/v2/manual.md#distribution-and-persistent-autocompletion)
- [zsh support](https://github.com/urfave/cli/blob/main/docs/v2/manual.md#zsh-support)
- [PowerShell support](https://github.com/urfave/cli/blob/main/docs/v2/manual.md#powershell-support)
