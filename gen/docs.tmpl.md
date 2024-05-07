# vai

## Usage

```sh
vai [task(s)] [flags]
```

```plaintext
{%.Flags%}
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
    date: ${{ input | fallback "now }}
```

```sh
vai echo --with name=$(whoami) --with date=$(date)
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

<!-- ## Task Schema

> Task name regex: `{%range $key, $value := .Schema.PatternProperties%}{% $key %}{%end%}`

<details>
<summary>View schema</summary>
```json
{%.SchemaJSON%}
```
</details> -->
