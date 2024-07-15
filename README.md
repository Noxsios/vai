# vai

![GitHub Tag](https://img.shields.io/github/v/tag/noxsios/vai)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/noxsios/vai)
[![codecov](https://codecov.io/gh/Noxsios/vai/graph/badge.svg?token=P7E9QC2RB9)](https://codecov.io/gh/Noxsios/vai)
[![Go Report Card](https://goreportcard.com/badge/github.com/noxsios/vai)](https://goreportcard.com/report/github.com/noxsios/vai)
![GitHub License](https://img.shields.io/github/license/noxsios/vai)
[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B45282%2Fgithub.com%2FNoxsios%2Fvai.svg?type=shield&issueType=license)](https://app.fossa.com/projects/custom%2B45282%2Fgithub.com%2FNoxsios%2Fvai?ref=badge_shield&issueType=license)

A simple task runner. Imagine GitHub actions and Makefile had a baby.

> [!CAUTION]
> This project is still in its early stages. Expect breaking changes.

## Installation

```sh
go install github.com/noxsios/vai/cmd/vai@latest
```

or if you like to live dangerously:

```sh
go install github.com/noxsios/vai/cmd/vai@main
```

## Example

![demo](https://github.com/user-attachments/assets/820f0cdc-99e3-413b-8fc8-c7b49dc5ad55)

Checkout more examples in the [docs](https://vai.razzle.cloud/docs/).

View CLI usage w/ `vai --help`

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

[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B45282%2Fgithub.com%2FNoxsios%2Fvai.svg?type=large&issueType=license)](https://app.fossa.com/projects/custom%2B45282%2Fgithub.com%2FNoxsios%2Fvai?ref=badge_large&issueType=license)
