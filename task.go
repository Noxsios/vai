package vai

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/invopop/jsonschema"
)

type Operation int

const (
	OperationRun Operation = iota
	OperationUses
)

type Task struct {
	CMD    *string `json:"cmd,omitempty"`
	Uses   *Uses   `json:"uses,omitempty"`
	With   `json:"with,omitempty"`
	Matrix `json:"matrix,omitempty"`
}

func (t Task) Operation() Operation {
	if t.CMD != nil {
		return OperationRun
	}
	if t.Uses != nil {
		return OperationUses
	}
	return -1
}

func (t Task) Run(with With) error {
	env := os.Environ()
	for k, v := range with {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd := exec.Command("sh", "-e", "-c", *t.CMD)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	logger.Print(*t.CMD)
	return cmd.Run()
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
		Required:   []string{"uses"},
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
