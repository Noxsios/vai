---
title: Workflow Syntax
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
    - run: go build -o bin/ -ldflags="-s -w" ./cmd/vai
      with:
        CGO_ENABLED: 0

test:
    - run: go test -v -race -cover -failfast -timeout 3m ./...

clean:
    - run: rm -rf bin/
```

  {{< /tab >}}

{{< /tabs >}}

## Task names

Task names must follow the following regex: `^[_a-zA-Z][a-zA-Z0-9_-]*$`. Try it out below:

<input spellcheck="false" placeholder="some-task" id="task-name-regex" />
<span id="regex-result"></span>

<script type="module" defer>
  const input = document.getElementById('task-name-regex');
  const result = document.getElementById('regex-result');
  input.addEventListener('input', () => {
    const regex = /^[_a-zA-Z][a-zA-Z0-9_-]*$/;
    if (input.value === '') {
      result.textContent = '';
      return;
    }
    const valid = regex.test(input.value);
    result.textContent = valid ? '✅' : '❌';
  });
</script>

### Examples of valid task names

```yaml
build: ...
another-task: ...
UPPERCASE: ...
mIxEdCaSe: ...
WithNumbers123: ...
```

## `eval` vs `run` vs `uses`

- `eval`: runs a [Tengo](https://github.com/d5/tengo) script
- `run`: runs a shell command/script
- `uses`: calls another task

All three can be used interchangeably within a task, and interoperate cleanly with `with`.

## Passing inputs

`with` follows a syntax similar to GitHub Actions expressions.

There are a few built-in functions that can be used in `with`, two shown below are:

- `input`: grabs the value passed to the task
  - If the task is top-level (called via CLI), `with` values are received from the `--with` flag.
  - If the task is called from another task, `with` values are passed from the calling step.
- `default`: sets a default value if the input is not provided

{{< tabs items="run,eval" >}}
{{< tab >}}

`with` is then mapped to the steps's environment variables, with key names being transformed to standard environment variable names (uppercase, with underscores).

```yaml {filename="vai.yaml"}
echo:
  - run: echo "Hello, $NAME, today is $DATE"
    with:
      name: ${{ input }}
      # default to "now" if not provided
      date: ${{ input | default "now" }}
```

```sh
vai echo --with name=$(whoami) --with date=$(date)
```

{{< /tab >}}
{{< tab >}}

`with` values are passed to the Tengo script as global variables at compilation time.

```yaml {filename="vai.yaml"}
echo:
  - eval: |
      fmt := import("fmt")
      if date == "" {
        times := import("times")
        date = times.time_format(times.now(), "2006-01-02")
      }
      s := fmt.sprintf("Hello, %s, today is %s", name, date)
      fmt.println(s)
    with:
      name: ${{ input }}
      # default to "now" if not provided
      date: ${{ input }}
```

```sh
vai echo --with name=$(whoami)
```

{{< /tab >}}
{{< /tabs >}}

## Run another task as a step

Calling another task within the same workflow is as simple as using the task name, similar to Makefile targets.

```yaml {filename="vai.yaml"}
general-kenobi:
  - run: echo "General Kenobi, you are a bold one"
  - run: echo "RESPONSE"
    with:
      response: ${{ input }}

hello:
  - run: echo "Hello There!"
  - uses: general-kenobi
    with:
      response: "Your move"
```

```sh
vai hello
```

## Run a task from a local file

Calling a task from a local file takes two arguments: the file path (required) and the task name (optional).

`file:<filepath>?task=<taskname>`

If the filepath is a directory, `vai.yaml` is appended to the path.

If the task name is not provided, the `default` task is run.

```yaml {filename="tasks/echo.yaml"}
simple:
  - run: echo "$MESSAGE"
    with:
      message: ${{ input }}
```

```yaml {filename="vai.yaml"}
echo:
  - uses: file:tasks/echo.yaml?task=simple
    with:
      message: ${{ input }}
```

```sh
vai echo --with message="Hello, World!"
```

## Run a task from a remote file

{{< callout emoji="⚠️" >}}
`uses` syntax leverages the [package-url spec](https://github.com/package-url/purl-spec)
{{< /callout >}}

{{< tabs items="GitHub,GitLab,HTTP(S)" >}}

{{< tab >}}

```yaml {filename="vai.yaml"}
remote-echo:
  - uses: pkg:github/noxsios/vai@main?task=echo#testdata/simple.yaml
    with:
      message: "Hello, World!"
```

{{< /tab >}}

{{< tab >}}

```yaml {filename="vai.yaml"}
remote-echo:
  - uses: pkg:gitlab/noxsios/vai@main?task=echo#testdata/simple.yaml
    with:
      message: "Hello, World!"
```

{{< /tab >}}

{{< tab >}}

```yaml {filename="vai.yaml"}
remote-echo:
  - uses: https://raw.githubusercontent.com/noxsios/vai/main/testdata/simple.yaml?task=echo
    with:
      message: "Hello, World!"
```

{{< /tab >}}

{{< /tabs >}}

```sh
vai remote-echo
```

## Persist variables between steps

Setting a variable with `persist` will persist it for the remaining steps in the task and can be overridden per-step.

```yaml {filename="vai.yaml",hl_lines=[4]}
set-name:
  - run: echo "Setting name to $NAME"
    with:
      name: ${{ input | persist }}
  - run: echo "Hello, $NAME"
  - run: echo "$NAME can be overridden per-step, but will persist between steps"
    with:
      name: "World"
  - run: echo "See? $NAME"
```

```sh
vai set-name --with name="Universe"
```

## Passing outputs

This leverages the same mechanism as GitHub Actions.

The `id` field is used to reference the output in subsequent steps.

The `from` function is used to reference the output from a previous step.

{{< tabs items="run,eval" >}}
{{< tab >}}

```yaml {filename="vai.yaml"}
color:
  - run: |
      echo "selected-color=green" >> $VAI_OUTPUT
    id: color-selector
  - run: echo "The selected color is $SELECTED"
    with:
      selected: ${{ from "color-selector" "selected-color" }}
```

{{< /tab >}}
{{< tab >}}

```yaml {filename="vai.yaml"}
color:
  - eval: |
      color := "green"
      vai_output["selected-color"] = color
    id: color-selector
  - run: echo "The selected color is $SELECTED"
    with:
      selected: ${{ from "color-selector" "selected-color" }}
```

{{< /tab >}}
{{< /tabs >}}

```sh
vai color
```
