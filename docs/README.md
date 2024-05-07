# vai

## Usage

```sh
vai [task(s)] [flags]
```

```plaintext
  -F, --force                 ignore checksum mismatch for cached remote files
      --list                  list available tasks
  -l, --log-level string      log level (default "info")
  -V, --version               print version
  -w, --with stringToString   variables to pass to the called task(s) (default [])

```

## Examples

<!-- TODO: auto gen this from tests -->

### List available tasks

```sh
# in a directory with a vai.yaml file
vai --list
```

### Run the "default" task

```sh
vai
```

### Run multiple tasks

```sh
vai task1 task2
```

### Run a task with variables

```yaml
# vai.yaml
echo:
  cmd: echo "Hello, $NAME, today is $DATE"
  with:
    name: ${{ input }}
    # default to "now" if not provided
    date: ${{ input | default "now" }}
```

```sh
vai echo --with name=$(whoami) --with date=$(date)
```

### Run another task as a step

```yaml
# vai.yaml
general-kenobi:
    cmd: echo "General Kenobi"

hello:
  cmd: echo "Hello There!"
  uses: general-kenobi
```

```sh
vai hello
```

### Run a task from a local file

```yaml
# tasks/echo.yaml
simple:
  cmd: echo $MESSAGE
  with:
    message: ${{ input }}
```

```yaml
# vai.yaml
echo:
  - uses: tasks/echo.yaml?task=simple
    with:
      message: ${{ input }}
```

```sh
vai echo --with message="Hello, World!"
```

### Run a task from a remote file

> `uses` syntax is an implementation of the package-url spec: `{url}@{version}?task={task}#{path}`

```yaml
# vai.yaml
remote-echo:
  uses: github.com/noxsios/vai@main?task=simple#tasks/echo.yaml
  with:
    message: "Hello, World!"
```

```sh
vai remote-echo
```

### Persist variables between steps

> NOTE: setting a variable with `persist` will persist it for the entire task
> and can be overridden per-task.
>
> This is not persistent between tasks. For that, pass the variable using `with`.

```yaml
# vai.yaml
set-name:
  cmd: echo "Setting name to $NAME"
  with:
    name: ${{ input | persist }}
  cmd: echo "Hello, $NAME"
  cmd: echo "$NAME can be overridden per-task, but will persist between tasks"
  with:
    name: "World"
  cmd: echo "See? $NAME"
```

```sh
vai set-name --with name="Universe"
```

### Passing outputs between steps

> This leverages the same mechanism as GitHub Actions.
>
> The `id` field is used to reference the output in subsequent steps.
>
> The `from` function is used to reference the output from a previous step.

```yaml
# vai.yaml
driving:
  - cmd: echo "Driving..."
  - cmd: |
      DESTINATION="Home"
      echo "Arrived at $DESTINATION"
      echo "destination:$DESTINATION" >> $VAI_OUTPUT
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

<!-- ## Task Schema

> Task name regex: `^[_a-zA-Z][a-zA-Z0-9_-]*$`

<details>
<summary>View schema</summary>
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://github.com/noxsios/vai/workflow",
  "$defs": {
    "Matrix": {
      "additionalProperties": {
        "items": true,
        "type": "array"
      },
      "type": "object"
    },
    "Step": {
      "oneOf": [
        {
          "properties": {
            "cmd": {
              "type": "string"
            },
            "uses": {
              "not": true
            }
          },
          "required": [
            "cmd"
          ]
        },
        {
          "properties": {
            "cmd": {
              "not": true
            },
            "uses": {
              "type": "string"
            }
          },
          "required": [
            "uses"
          ]
        }
      ],
      "properties": {
        "cmd": {
          "type": "string",
          "description": "Command to run"
        },
        "uses": {
          "type": "string",
          "description": "Location of a remote task to call conforming to the purl spec"
        },
        "id": {
          "type": "string",
          "description": "Unique identifier for the step"
        },
        "description": {
          "type": "string",
          "description": "Description of the step"
        },
        "with": {
          "patternProperties": {
            "^[a-zA-Z_]+[a-zA-Z0-9_]*$": {
              "oneOf": [
                {
                  "type": "string"
                },
                {
                  "type": "boolean"
                },
                {
                  "type": "integer"
                }
              ]
            }
          },
          "additionalProperties": false,
          "type": "object",
          "minItems": 1,
          "description": "Additional parameters for the step/task call"
        },
        "matrix": {
          "additionalProperties": {
            "items": {
              "oneOf": [
                {
                  "type": "string"
                },
                {
                  "type": "boolean"
                },
                {
                  "type": "integer"
                }
              ]
            },
            "type": "array"
          },
          "type": "object",
          "minItems": 1,
          "description": "Matrix of parameters"
        }
      },
      "additionalProperties": false,
      "type": "object"
    },
    "Task": {
      "items": {
        "$ref": "#/$defs/Step"
      },
      "type": "array"
    },
    "With": {
      "type": "object"
    }
  },
  "patternProperties": {
    "^[_a-zA-Z][a-zA-Z0-9_-]*$": {
      "$ref": "#/$defs/Task",
      "description": "Name of the task"
    }
  },
  "additionalProperties": false,
  "type": "object"
}
```
</details> -->
