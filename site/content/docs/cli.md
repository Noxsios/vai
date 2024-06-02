---
title: CLI
---

## Usage

```bash
{{< usage >}}
```

## Discover available tasks

The `--list` flag can be used to list all the tasks in a Vai workflow.

If defined, the `default` task will be listed first. Otherwise, tasks will be listed in alphabetical order.

```sh
$ vai --list

Available:

- default
- build
- test
```

## "default" task

The task named `default` in a Vai workflow is the default task that will be run when no task is specified.

```sh
$ vai
# is equivalent to
$ vai default
```

## Run multiple tasks

Like `make`, you can run multiple tasks in a single command.

```sh
$ vai task1 task2
```

## Specify a workflow file

By default, Vai will look for a file named `vai.yaml` in the current directory. You can specify a different file to use with the `--file` or `-f` flag.

```sh
$ vai --file path/to/other.yaml
```

## Shell completions

Like `make`, `vai` only has a single command. As such, shell completions are not generated in the normal way most Cobra CLI applications are (i.e. `vai completion bash`). Instead, you can use the following snippet to generate completions for your shell:

{{< tabs items="bash,zsh,fish,powershell" >}}

{{< tab "bash" >}}

```bash
VAI_COMPLETION=true vai completion bash
```

{{< /tab >}}

{{< tab "zsh" >}}

```zsh
VAI_COMPLETION=true vai completion zsh
```

{{< /tab >}}

{{< tab "fish" >}}

```fish
VAI_COMPLETION=true vai completion fish
```

{{< /tab >}}

{{< tab "powershell" >}}

```powershell
$env:VAI_COMPLETION='true'; vai completion powershell; $env:VAI_COMPLETION=$null
```

{{< /tab >}}

{{< /tabs >}}

Completions are only generated when the `VAI_COMPLETION` environment variable is set to `true`, and the `completion <shell>` arguments are passed to the `vai` command.

This is because `completion bash|fish|etc...` are valid task names in a Vai workflow, so the CLI would attempt to run these tasks. By setting the environment variable, the CLI knows to generate completions instead of running tasks.
