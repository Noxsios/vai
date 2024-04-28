package vai

import (
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"strings"

	"github.com/invopop/jsonschema"
)

type Task struct {
	CMD  *string `json:"cmd,omitempty"`
	Uses *string `json:"uses,omitempty"`
	With With    `json:"with,omitempty"`
}

func (t Task) Run() error {
	env := os.Environ()
	for k, v := range t.With {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd := exec.Command("sh", "-e", "-c", *t.CMD)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println(*t.CMD)
	return cmd.Run()
}

type With map[string]interface{}

func (w With) StringMap() (map[string]string, error) {
	out := make(map[string]string)
	for k, v := range w {
		switch v := v.(type) {
		case string:
			out[k] = v
		case bool:
			out[k] = fmt.Sprintf("%t", v)
		case int:
			out[k] = fmt.Sprintf("%d", v)
		default:
			return nil, fmt.Errorf("invalid type for key %s", k)
		}
	}
	return out, nil
}

func (w With) FromStringMap(m map[string]string) {
	for k, v := range m {
		w[k] = v
	}
}

func (w With) Set(k string, v any) {
	w[k] = v
}

func (w With) Apply(parent With) error {
	tmpl := template.New("with and inputs").Option("missingkey=error").Delims("${{", "}}")

	fmt.Println("# templating", w, "with", parent)

	for k, v := range w {
		val := fmt.Sprintf("%s", v)
		fm := template.FuncMap{
			"input": func(fallback ...any) any {
				v, ok := parent[k]
				if !ok || v == "" {
					if len(fallback) == 0 {
						return nil
					}
					return fallback[0]
				}
				return v
			},
		}
		tmpl.Funcs(fm)
		tmpl, err := tmpl.Parse(val)
		if err != nil {
			return err
		}
		var templated strings.Builder
		if err := tmpl.Execute(&templated, nil); err != nil {
			return err
		}
		result := templated.String()
		w[k] = result
	}
	return nil
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
	fmt.Println(":", call)
	tg, ok := wf[call]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return tg, nil
}
