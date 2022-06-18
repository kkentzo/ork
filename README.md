[![test](https://github.com/kkentzo/ork/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/kkentzo/ork/actions/workflows/test.yml)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/kkentzo/ork)

# ork

`ork` is a program to organize, compose and execute command workflows
in software projects.

It is meant as a modern, light-weight substitute to the venerable
`Makefile` program and especially for `Makefile`s that make heavy use
of `.PHONY` targets (i.e. `Makefile`s that focus on operational tasks
rather than files).

`ork` is simple and mostly inspired by the workflow syntax of
[Github](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)
and [Gitlab](https://docs.gitlab.com/ee/ci/yaml/gitlab_ci_yaml.html)
CI products.

## Features

### Workflow organization

`ork` organizes workflows in yaml format (`Orkfile.yml`) as a
collection of tasks, each consisting of a sequence of actions that are
executed independently of each other (i.e. in separate processes). If
an action results in an error, then the chain stops.

Here's an example:

```yaml
global:
  default: build
  env:
    - GO_BUILD: go -ldflags="-s -w"

tasks:

  - name: build
    description: build the application
    env:
      - GOOS: linux
        GOARCH: amd64
        GO_TARGET: bin/foo
    actions:
      - $GO_BUILD -o $GO_TARGET

```

A task can contain any kind of executable actions, e.g.:

```yaml
tasks:
  - name: say/hello
    description: Say hello from python
    actions:
      - python -c "import sys; sys.stdout.write('hello from python')"
```

A task can also contain an arbitrary number of nested tasks for
grouping together related functionalities, for example:

```yaml
tasks:
  - name: db
    description: task group for various database-related actions
    env:
      - DB_CONN: "postgres://..."
    tasks:
      - name: migrate
        actions:
          # commands for applying database migrations
      - name: rollback
        actions:
          # commands for applying database migrations
      ...
```

The above configuration will generate three tasks: `db`, `db.migrate`
and `db.rollback`. Note that the parent task (`db`) is still
considered a task whose environment, actions, hooks etc. will be
executed before its children (so that the parent task can be used for
setting up the children tasks).

### Global configuration

As seen above, Orkfiles can also have a global section for setting up
properties for all tasks such as environment variables, the default
task etc. Global environment variables are overriden by local
(task-specific) ones, e.g.:

```yaml
global:
  env:
    - VAR: bar

  tasks:
    - name: foo
      env:
        - VAR: foo
      actions:
        - echo $VAR
```

In this case, `ork foo` will output `foo` (the task-local version of
`$VAR`).

### Environment Variables

Environment variables (global and task-specific) support command
substitution interpolation using the special form `$[...]` like so:

```yaml
tasks:

  - name: foobar
    env:
      - VAR: $[echo foo]-$[echo bar]
    actions:
      - echo $VAR
```

The output of running `ork foobar` on the above Orkfile will be
`foo-bar`.

Environment variables can be defined in different groups so that each
group can utilize values from the previous group. The following
example will output `a`:

```yaml
tasks:
  - name: foo
    env:
      - A: a
      - B: $[bash -c "echo $a"]
    actions:
      - echo $B
```

Environment variables are expanded before the action is actually
executed. In the example above, this means that `$VAR` will be
replaced by its value `foo-bar`, so the action that will be executed
will be `echo foo-bar`. This behaviour can be disabled (using
`expand_env: false` in task) so that actions that set their own
variables can be correctly executed like in the following example:

```yaml
tasks:
  - name: foo
    expand_env: false
    actions:
      - bash -c "for f in $(ls -1 db/seeds/*.sql); do echo $f; done; "
```

If `expand_env` was true (the default behaviour), then the token `$f`
would be substituted with an empty string before the action was
executed and we would not receive the expected output from the
command.

The matching of the substitution pattern `$[...]` can be problematic
for statements like

`$[bash -c "if [ \"${DEPLOY_ENV}\" == \"production\" ]; then echo production; else echo staging; fi"]`

In this case, the opening `$[` will be matched in a non-greedy manner
by the closing `]` of the bash `if` statement and the command will
fail. `ork` exposes a task attribute called `env_subst_greedy`
(default: false) which can be used to enforce the desired behaviour
(in this case it must be set to true).

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

### Task Success/Error Hooks

Orkfiles support post-action hooks for individual tasks, e.g.:

```yaml
tasks:
  - name: deploy
    env:
      - RELEASE_TAG: release/$[date '+%Y-%m-%dT%H-%M-%S']
    actions:
      - ...
    on_success:
      - git tag -a $RELEASE_TAG -m "$RELEASE_TAG"
    on_failure:
      - curl -d '{"error":"$ORK_ERROR"}' -H "Content-Type: application/json" -X POST http://notifications.somewhere
```

The `on_success` action hooks will be executed only if the task action
chain is executed successfully. In the event of an error,

If all of the task's actions are completed without any errors, then
the `on_success` actions are executed, otherwise the `on_failure`
actions are executed with access to the `$ORK_ERROR` environment
variable.

### Working directory

A task can specify its own working directory like so:

```yaml
tasks:
  - name: deploy
    working_dir: ./ansible
    actions:
      - ansible-playbook -i hosts app.yml
```

All the task's actions will have `./ansible` as their working
directory.

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
