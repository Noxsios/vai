package vai

import (
	"github.com/invopop/jsonschema"
)

// Operation is an enum for the type of operation a step is performing
type Operation int

const (
	// OperationUnknown is an unknown operation
	OperationUnknown Operation = iota
	// OperationRun is a step that runs a command
	OperationRun
	// OperationUses is a step that calls another task
	OperationUses
)

// Step is a single step in a task
//
// While a step can have both `cmd` and `uses` fields, only one of them can be set
// at a time.
//
// This is enforced by JSON schema validation.
//
// TODO:
// - add `if` and `continue-on-error` fields?
// - add `timeout` field?
type Step struct {
	// CMD is the command to run
	CMD string `json:"cmd,omitempty"`
	// Uses is a reference to a remote task
	Uses string `json:"uses,omitempty"`
	// With is a map of additional parameters for the step/task call
	With `json:"with,omitempty"`
	// Matrix is a matrix of parameters to run the step/task with
	Matrix `json:"matrix,omitempty"`
	// ID is a unique identifier for the step
	ID string `json:"id,omitempty"`
	// Description is a description of the step
	Description string `json:"description,omitempty"`
}

// Operation returns the type of operation the step is performing
func (s Step) Operation() Operation {
	if s.CMD != "" {
		return OperationRun
	}
	if s.Uses != "" {
		return OperationUses
	}
	return OperationUnknown
}

// JSONSchemaExtend extends the JSON schema for a step
func (Step) JSONSchemaExtend(schema *jsonschema.Schema) {
	not := &jsonschema.Schema{
		Not: &jsonschema.Schema{},
	}

	props := jsonschema.NewProperties()
	props.Set("cmd", &jsonschema.Schema{
		Type:        "string",
		Description: "Command to run",
	})
	props.Set("uses", &jsonschema.Schema{
		Type:        "string",
		Description: "Location of a remote task to call conforming to the purl spec",
	})
	props.Set("id", &jsonschema.Schema{
		Type:        "string",
		Description: "Unique identifier for the step",
	})
	props.Set("description", &jsonschema.Schema{
		Type:        "string",
		Description: "Description of the step",
	})

	oneOfStringIntBool := &jsonschema.Schema{
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
	}

	var single uint64 = 1

	with := &jsonschema.Schema{
		Type:                 "object",
		Description:          "Additional parameters for the step/task call",
		MinItems:             &single,
		AdditionalProperties: oneOfStringIntBool,
	}

	props.Set("with", with)

	matrix := &jsonschema.Schema{
		Type:        "object",
		Description: "Matrix of parameters",
		MinItems:    &single,
		AdditionalProperties: &jsonschema.Schema{
			Type:  "array",
			Items: oneOfStringIntBool,
		},
	}

	props.Set("matrix", matrix)

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

	schema.Properties = props
	schema.OneOf = []*jsonschema.Schema{
		oneOfCmd,
		oneOfUses,
	}
}

