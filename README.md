[![test](https://github.com/kkentzo/ork/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/kkentzo/ork/actions/workflows/test.yml)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/kkentzo/ork)

# ork

`ork` is a program to organize, compose and execute command workflows
in software projects.

It is meant as a modern, light-weight substitute to the venerable
`Makefile` program and especially for Makefiles that make heavy use of
`.PHONY` targets (i.e. Makefiles that focus on the orchestration of
operations rather than the generation of target files).

`ork` aims to stay simple and is mostly inspired by the workflow
syntax of
[Github](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)
and [Gitlab](https://docs.gitlab.com/ee/ci/yaml/gitlab_ci_yaml.html)
CI products.

## Features

### Workflow organization

`ork` organizes workflows in yaml format (`Orkfile.yml`) as a
hierarchy of tasks, with each task having the following optional
characteristics:

- a sequence of `actions` that are executed independently of each other
(i.e. in separate processes)
- a sequence of task dependencies (`depends_on`) that serve as
  prerequisites for the current task
- a set of task-specific environment variables (`env`)
- post-execution hooks (`on_success`, `on_error`)

The execution of any task is preceded by the execution of all the
tasks in the hierarchy chain in top-down order.

Here's a simple example:

```yaml
default: build

tasks:

  - name: build
    description: build the application
    env:
      - GOOS: linux
        GOARCH: amd64
      - GO_TARGET: bin/foo
        GO_BUILD: go -ldflags="-s -w"
    actions:
      - $GO_BUILD -o $GO_TARGET

  - name: test
    description: test the application
    depends_on:
      - build
    actions:
      - go test . -v -cover -count=1

```

In the example above there are two tasks: `build` and `test`. Running
`ork` on the Orkfile will execute the default `build` task. Running
`ork test` will execute first the `build` task and then the `test`
task.

A task can contain any kind of executable actions, e.g.:

```yaml
tasks:
  - name: say.hello
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

### Environment Variables

#### Command substitution

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

#### Variable expansion

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

#### Variable grouping and ordering

The ordering of environment variables in a single group is random:

```yaml
tasks:
  - name: foo
    env:
      - A: a
        B: ${A}
    actions:
      - echo $B
```

The output in this case would randomly vary between `a` and ``.

However, environment variables can be defined in different (ordered)
groups so that each group can utilize values from the previous
group. The following example will always output `a`:

```yaml
tasks:
  - name: foo
    env:
      - A: a
      - B: ${A}
    actions:
      - echo $B
```

#### Command substitution pattern matching

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

### Task Requirements

Tasks can express requirements in terms of the environment variables
that should be present and/or have specific expected values. If a
requirement is not met, then the task stops with an error.

Requirements can be expressed in two ways.

The first is to require the existence of an environment variable like
so:

```yaml
tasks:
  - name: a
    require:
      exists:
        - A
    actions:
      - echo $A
```

Task `a` will fail, since the environment variable `A` is not defined
when task `a` is executed. In contrast, given the following Orkfile,
task `b` will succeed:

```yaml
tasks:
  - name: a
    env:
      - A: a
  - name: b
    depends_on:
      - a
    require:
      exists:
        - A
    actions:
      - echo $A
```

Tasks can also specify requirements in terms of the expected value of
environment variables:

```yaml
tasks:
  - name: a
    env:
      - A: a
  - name: b
    depends_on:
      - a
    require:
      equals:
        A: foo
    actions:
      - echo $A
```

Task `b` will fail since, when executed, the environment variable `A`
has the value `a` instead of the expected value `foo`.

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

### Dynamic Task Generation

Dynamic tasks that are generated at runtime can be defined in the
Orkfile as follows:

```yaml
tasks:
  - name: deploy
    env:
      - TOP_LEVEL_ENV: deploy
    generate:
      - name: production
        env:
          - SERVER_URL: http://i_am_production
        actions:
          - echo $SERVER_URL
      - name: staging
        env:
          - SERVER_URL: http://i_am_staging
        actions:
          - echo $SERVER_URL
    tasks:
      - name: ping
        actions:
          - echo "${TOP_LEVEL_ENV} => pinging ${SERVER_URL}"
      - name: send
        actions:
          - echo "${TOP_LEVEL_ENV} => sending to ${SERVER_URL}"
```

Dynamic tasks are created during runtime as an extra layer between the
current task and its nested tasks. The above Orkfile defines two
dynamic tasks (`production` and `staging`) under which both the `ping`
and `send` tasks will be available, so the following actionable tasks
will be constructed:

- `deploy.production.ping`
- `deploy.production.send`
- `deploy.staging.ping`
- `deploy.staging.send`

The rules of nested tasks still apply, i.e. tasks are executed
top-down (parent to children) and all the typical task characteristics
(`env`, `actions`, `on_success`) are available for dynamic tasks.

So, if we execute `ork deploy.staging.ping`, the output will be:
`deploy => pinging http://i_am_staging`.

## Installation & Usage

`ork` can be installed by downloading the latest release binary from
[here](https://github.com/kkentzo/ork/releases) according to your
platform. There are different binaries supporting intel (amd64) or arm
architectures for linux or darwin (mac). The downloaded binary can
then be installed in a convenient place in your `$PATH` (e.g. under
`/usr/local/bin`).

`ork` can execute one or more tasks defined in an Orkfile by running:

```bash
$ ork task1 task2 ...
```

Run `ork -h` for program options.

## Autocompletion

`ork` supports task autocompletion in the command-line. Follow the
guides below to enable system-wide auto-completion (courtesy of the
excellent [cli](https://github.com/urfave/cli) library):

- [bash instructions](https://github.com/urfave/cli/blob/main/docs/v2/examples/bash-completions.md)
- [zsh support](https://github.com/urfave/cli/blob/main/docs/v2/examples/bash-completions.md#zsh-support)
- [PowerShell support](https://github.com/urfave/cli/blob/main/docs/v2/examples/bash-completions.md#powershell-support)
