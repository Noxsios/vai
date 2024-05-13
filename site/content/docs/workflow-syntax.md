---
title: 'Workflow Syntax'
---

A Vai workflow is any YAML file that conforms to the [`vai` schema](./schema-validation.md#raw-schema).

Unless specified, the default file name is `vai.yaml`.

## Structure

Similar to `Makefile`s, a Vai workflow is a map of tasks, where each task is a series of steps.

Checkout the comparison below:

{{< tabs items="Makefile,Vai" >}}

  {{< tab >}}

```makefile {filename="Makefile"}
.DEFAULT_GOAL := build

build:
	CGO_ENABLED=0 go build -o bin/ -ldflags="-s -w" ./cmd/vai

test:
	go test -v -race -cover -failfast -timeout 3m ./...

clean:
	rm -rf bin/
```

  {{< /tab >}}
  {{< tab >}}

```yaml {filename="vai.yaml"}
default:
    - uses: build

build:
    - cmd: CGO_ENABLED=0 go build -o bin/ -ldflags="-s -w" ./cmd/vai

test:
    - cmd: go test -v -race -cover -failfast -timeout 3m ./...

clean:
    - cmd: rm -rf bin/
```

  {{< /tab >}}

{{< /tabs >}}

## Task names

Task names must follow the following regex: `^[_a-zA-Z][a-zA-Z0-9_-]*$`.

### Examples of valid task names

```yaml
build: ...
another-task: ...
UPPERCASE: ...
mIxEdCaSe: ...
WithNumbers123: ...
```

## `--list` flag

The `--list` flag can be used to list all the tasks in a Vai workflow.

If defined, the `default` task will be listed first. Otherwise, tasks will be listed in alphabetical order.

### Example of listing tasks

```sh
$ vai --list

Available:

- default
- build
- test
```

## Run the "default" task

The task named `default` in a Vai workflow is the default task that will be run when no task is specified.

```sh
$ vai
# is equivalent to
$ vai default
```

## Run multiple tasks

Like `Makefile`, you can run multiple tasks in a single command.

```sh
vai task1 task2
```

## Run a task with variables

`with` follows a syntax similar to GitHub Actions expressions.

There are a few built-in functions that can be used in `with`, two shown below are:

- `input`: grabs the value passed to the task
  - If the task is top-level (called via CLI), `with` values are received from the `--with` flag.
  - If the task is called from another task, `with` values are passed from the calling step.
- `default`: sets a default value if the input is not provided

The `with` map is then mapped to the steps's environment variables, with key names being transformed to standard environment variable names (uppercase, with underscores).

```yaml {filename="vai.yaml"}
echo:
  - cmd: echo "Hello, $NAME, today is $DATE"
    with:
      name: ${{ input }}
      # default to "now" if not provided
      date: ${{ input | default "now" }}
```

```sh
vai echo --with name=$(whoami) --with date=$(date)
```

## Run another task as a step

Calling another task within the same file is as simple as using the task name.

```yaml {filename="vai.yaml"}
general-kenobi:
  - cmd: echo "General Kenobi"

hello:
  - cmd: echo "Hello There!"
  - uses: general-kenobi
```

```sh
vai hello
```

## Run a task from a local file

Calling a task from a local file takes two arguments: the file path (required) and the task name (optional).

`<filepath>?task=<taskname>`

If the filepath is a directory, `vai.yaml` is appended to the path.

If the task name is not provided, the `default` task is run.

```yaml {filename="tasks/echo.yaml"}
simple:
  - cmd: echo "$MESSAGE"
    with:
      message: ${{ input }}
```

```yaml {filename="vai.yaml"}
echo:
  - uses: tasks/echo.yaml?task=simple
    with:
      message: ${{ input }}
```

```sh
vai echo --with message="Hello, World!"
```

## Run a task from a remote file

> [!WARNING]
> Currently only supports GitHub repos.
>
> `uses` syntax leverages the package-url spec: `{url}@{version}?task={task}#{path}`

```yaml {filename="vai.yaml"}
remote-echo:
    # run the "simple" task from the "tests/testdata/echo.yaml" file in the "github.com/noxsios/vai" repo on the "main" branch
  - uses: github.com/noxsios/vai@main?task=simple#tests/testdata/echo.yaml
    with:
      message: "Hello, World!"
```

```sh
vai remote-echo
```

## Persist variables between steps

Setting a variable with `persist` will persist it for the remaining steps in the task
and can be overridden per-step.

```yaml {filename="vai.yaml"}
set-name:
  - cmd: echo "Setting name to $NAME"
    with:
      name: ${{ input | persist }}
  - cmd: echo "Hello, $NAME"
  - cmd: echo "$NAME can be overridden per-step, but will persist between steps"
    with:
      name: "World"
  - cmd: echo "See? $NAME"
```

```sh
vai set-name --with name="Universe"
```

## Passing outputs between steps

> This leverages the same mechanism as GitHub Actions.
>
> The `id` field is used to reference the output in subsequent steps.
>
> The `from` function is used to reference the output from a previous step.

```yaml {filename="vai.yaml"}
driving:
  - cmd: echo "Driving..."
  - cmd: |
      DESTINATION="Home"
      echo "Arrived at $DESTINATION"
      echo "destination=$DESTINATION" >> $VAI_OUTPUT
    id: history    
  - cmd: |
      echo "Done driving"
      echo "I arrived at $LOCATION"
    with:
      location: ${{ from "history" "destination" }}
```

```sh
vai driving
```
