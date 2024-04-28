# vai

A simple task runner. Imagine GitHub actions and Makefile had a baby.

## Installation

```sh
go install github.com/Noxsios/vai@latest
```

## Example

```bash
cat <<EOF > vai.yaml
echo:
  - cmd: echo "\$message"
    with:
      message: \${{ input }}
EOF
```

```sh
$ vai echo --with message="Hello World!"
hello world
```

Learn more w/ `vai --help`

## Schema Validation

Enabling schema validation in VSCode:

```json
    "yaml.schemas": {
        "https:///raw.githubusercontent.com/Noxsios/vai/main/vai.schema.json": "vai.yaml",
    },
```

Per file basis:

```yaml
# yaml-language-server: $schema=https:///raw.githubusercontent.com/Noxsios/vai/main/vai.schema.json
```
