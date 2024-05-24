---
title: Schema Validation
---

Enabling schema validation in VSCode:

```json {filename=".vscode/settings.json"}
    "yaml.schemas": {
        "https://raw.githubusercontent.com/noxsios/vai/main/vai.schema.json": "vai.yaml",
    },
```

Per file basis:

```yaml {filename="some-task.yaml"}
# yaml-language-server: $schema=https://raw.githubusercontent.com/noxsios/vai/main/vai.schema.json
```

## Raw Schema

```plaintext {filename="vai.schema.json"}
{{< schema >}}
```
