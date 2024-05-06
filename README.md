# vai

![GitHub Tag](https://img.shields.io/github/v/tag/noxsios/vai)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/noxsios/vai)
[![codecov](https://codecov.io/gh/Noxsios/vai/graph/badge.svg?token=P7E9QC2RB9)](https://codecov.io/gh/Noxsios/vai)
![GitHub License](https://img.shields.io/github/license/noxsios/vai)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/noxsios/vai)

A simple task runner. Imagine GitHub actions and Makefile had a baby.

> [!CAUTION]
> This project is still in its early stages. Expect breaking changes.

## Installation

```sh
go install github.com/noxsios/vai/cmd/vai@latest
```

To update to the latest version:

```sh
rm $(which vai)
go install github.com/noxsios/vai/cmd/vai@latest
```

## Example

```bash
cat <<EOF > vai.yaml
echo:
  - cmd: echo "\$MESSAGE"
    with:
      message: \${{ input }}

remote-echo:
  - uses: github.com/noxsios/vai@main?task=simple#tasks/echo.yaml
    with:
      message: hello from main
EOF
```

```sh
$ vai echo --with message="Hello World!"
echo "$MESSAGE"

Hello World!
```

Learn more w/ `vai --help`

## Schema Validation

Enabling schema validation in VSCode:

```json
    "yaml.schemas": {
        "https://raw.githubusercontent.com/noxsios/vai/main/vai.schema.json": "vai.yaml",
    },
```

Per file basis:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/noxsios/vai/main/vai.schema.json
```
