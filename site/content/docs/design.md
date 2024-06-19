---
title: Design
---

At the end of the day, Vai's objective is to orchestrate shell commands from local and remote sources in a sane manner.

Vai values:

- Simplicity of configuration and usage
- Leveraging and extending existing systems over defining new ones

## Inspiration from `make`

Drawing inspiration from `make`, Vai only has a single command: `vai`. Any sub-commands are parsed as task calls (e.g. `vai hello world` is the equivalent of `make hello world`). Vai will also look for a `vai.yaml` in the current working directory and error unless a path is passed via the `-f|--file` flag.

Not placing task calls behind a `run` subcommand was a deliberate choice. In my mind, creating subcommands would be a slippery slope that would lead to scope creep.

I decided to use Cobra's `StringToStringVarP` to pass arguments to the called task(s) instead of replicating `make`'s variables syntax. `vai hello --with text=world --with debug=true` versus `make hello TEXT=world DEBUG=true`. While more verbose, I felt this reads a little better.

> I also borrowed the concept of the `.DEFAULT_GOAL` for Vai's [../cli#default-task]

Another major design decision I borrowed from `make` was not mutating the current working directory at all. Given the following repository structure:

{{< filetree/container >}}
  {{< filetree/folder name="tasks" >}}
    {{< filetree/file name="vai.yaml" >}}
  {{< /filetree/folder >}}
  {{< filetree/folder name="cmd" >}}
    {{< filetree/folder name="vai" >}}
      {{< filetree/file name="main.go" >}}
    {{< /filetree/folder >}}
  {{< /filetree/folder >}}
{{< /filetree/container >}}

```yaml {filename="tasks/vai.yaml"}
default:
  - uses: build

build:
  - run: CGO_ENABLED=0 go build -o bin/ -ldflags="-s -w" ./cmd/vai
```

The following works when run from the root of the repo:

```bash
vai -f tasks/vai.yaml
```

In Vai, all tasks are run from the context of the current working directory, if a task needs to run in a specific directory, it should call `cd` from within.

## Inspiration from GitHub Actions

The majority of Vai's [workflow schema](../schema-validation#raw-schema) was either inspired by, or is a direct replication of [GitHub Workflow's JSON schema](https://github.com/SchemaStore/schemastore/blob/master/src/schemas/json/github-workflow.json).

GitHub orchestrates `jobs`, which are collections of either `run` or `uses` steps.

Vai orchestrates `tasks`, which are collections of either `run` or `uses` steps.

The main differences between these two:

- Vai's tasks are top-level in the YAML definition
- Vai's tasks have no configuration outside of their definition

```yaml {filename="vai.yaml"}
default:
  - uses: build

build:
  - run: CGO_ENABLED=0 go build -o bin/ -ldflags="-s -w" ./cmd/vai
```

> Indentation is one of the largest contributing factors to cognitive complexity when it comes to reading code. I like YAML, but have found it can become unwieldy once you get more than 3 indentation layers deep. Vai strives to remain within this 3 layer rule.

## Import System

- github.com/package-url/packageurl-go
- caching
- gitlab vs github API for SHA
- relative pathing from `file:` from remote sources

## Testing

- testscript
- testify
- fuzzing
- coverage as a goal
