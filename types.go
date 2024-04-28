package vai

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/invopop/jsonschema"
)

type Task struct {
	CMD    *string `json:"cmd,omitempty"`
	Uses   *string `json:"uses,omitempty"`
	With   With    `json:"with,omitempty"`
	Matrix Matrix  `json:"matrix,omitempty"`
}

type Matrix map[string][]interface{}

type MatrixInstance map[string]interface{}

func (t Task) Run(with With) error {
	env := os.Environ()
	for k, v := range with {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd := exec.Command("sh", "-e", "-c", *t.CMD)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type With map[string]interface{}

func PeformLookups(parent, child With, mi MatrixInstance) (With, error) {
	logger.Debug("templating", "parent", parent, "child", child, "instance", mi)

	r := make(With)

	for k, v := range child {
		val := fmt.Sprintf("%s", v)
		fm := template.FuncMap{
			"input": func(fallback ...any) any {
				v, ok := parent[k]
				if !ok || v == "" {
					if len(fallback) == 0 {
						logger.Warn("no input", "key", k)
						return nil
					}
					return fallback[0]
				}
				return v
			},
			"matrix": func(m string) string {
				if mi == nil {
					logger.Warn("no matrix instance", "key", m)
					return ""
				}
				v, ok := mi[m]
				if !ok {
					logger.Warn("no matrix value", "key", m)
					return ""
				}
				return fmt.Sprintf("%v", v)
			},
		}
		tmpl := template.New("withs").Option("missingkey=error").Delims("${{", "}}")
		tmpl.Funcs(fm)
		tmpl, err := tmpl.Parse(val)
		if err != nil {
			return nil, err
		}
		var templated strings.Builder
		if err := tmpl.Execute(&templated, nil); err != nil {
			return nil, err
		}
		result := templated.String()
		r[k] = result
	}
	logger.Debug("templated", "result", r)
	return r, nil
}

func (t Task) JSONSchema() *jsonschema.Schema {
	props := jsonschema.NewProperties()
	not := &jsonschema.Schema{
		Not: &jsonschema.Schema{},
	}
	props.Set("cmd", &jsonschema.Schema{
		Type:        "string",
		Description: "Command to run",
	})
	props.Set("uses", &jsonschema.Schema{
		Type:        "string",
		Description: "Call another task",
	})

	with := &jsonschema.Schema{
		Type:        "object",
		Description: "Additional parameters for the task",
		AdditionalProperties: &jsonschema.Schema{
			OneOf: []*jsonschema.Schema{
				{
					Type: "string",
				},
				{
					Type: "boolean",
				},
				{
					Type: "integer",
				},
			},
		},
	}

	matrix := &jsonschema.Schema{
		Type:        "object",
		Description: "Matrix of parameters",
		AdditionalProperties: &jsonschema.Schema{
			Type: "array",
			Items: &jsonschema.Schema{
				OneOf: []*jsonschema.Schema{
					{
						Type: "string",
					},
					{
						Type: "boolean",
					},
					{
						Type: "integer",
					},
				},
			},
		},
	}

	props.Set("matrix", matrix)

	props.Set("with", with)

	cmdProps := jsonschema.NewProperties()
	cmdProps.Set("cmd", &jsonschema.Schema{
		Type: "string",
	})
	cmdProps.Set("uses", not)
	oneOfCmd := &jsonschema.Schema{
		Required:   []string{"cmd"},
		Properties: cmdProps,
	}

	usesProps := jsonschema.NewProperties()
	usesProps.Set("cmd", not)
	usesProps.Set("uses", &jsonschema.Schema{
		Type: "string",
	})
	oneOfUses := &jsonschema.Schema{
		Required:   []string{"uses", "with"},
		Properties: usesProps,
	}

	s := &jsonschema.Schema{
		Type:                 "object",
		Properties:           props,
		AdditionalProperties: jsonschema.FalseSchema,
		OneOf: []*jsonschema.Schema{
			oneOfCmd,
			oneOfUses,
		},
	}

	return s
}

type TaskGroup []Task

type Workflow map[string]TaskGroup

func (wf Workflow) Find(call string) (TaskGroup, error) {
	logger.Debug("finding", "task", call)
	tg, ok := wf[call]
	if !ok {
		return nil, fmt.Errorf("task group %q not found", call)
	}
	return tg, nil
}
