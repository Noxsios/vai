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

![demo](https://github.com/Noxsios/vai/assets/50058333/c0c1e906-deb1-4601-814b-8e36a4e8b322)

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
