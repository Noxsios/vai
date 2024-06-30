// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"github.com/invopop/jsonschema"
)

// Step is a single step in a task
//
// While a step can have both `run` and `uses` fields, only one of them can be set
// at a time.
//
// This is enforced by JSON schema validation.
type Step struct {
	// Run is the command/script to run
	Run string `json:"run,omitempty"`
	// Eval is an expression to evaluate with tengo
	Eval string `json:"eval,omitempty"`
	// Uses is a reference to another task
	Uses string `json:"uses,omitempty"`
	// With is a map of additional parameters for the step/task call
	With `json:"with,omitempty"`
	// ID is a unique identifier for the step
	ID string `json:"id,omitempty"`
	// Name is a human-readable name for the step
	Name string `json:"name,omitempty"`
}

// JSONSchemaExtend extends the JSON schema for a step
func (Step) JSONSchemaExtend(schema *jsonschema.Schema) {
	not := &jsonschema.Schema{
		Not: &jsonschema.Schema{},
	}

	props := jsonschema.NewProperties()
	props.Set("run", &jsonschema.Schema{
		Type:        "string",
		Description: "Command/script to run",
	})
	props.Set("uses", &jsonschema.Schema{
		Type:        "string",
		Description: "Location of a remote task to call conforming to the purl spec",
	})
	props.Set("eval", &jsonschema.Schema{
		Type:        "string",
		Description: "Expression to evaluate with tengo",
	})
	props.Set("id", &jsonschema.Schema{
		Type:        "string",
		Description: "Unique identifier for the step",
	})
	props.Set("name", &jsonschema.Schema{
		Type:        "string",
		Description: "Human-readable name for the step",
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
		Type:        "object",
		Description: "Additional parameters for the step/task call",
		MinItems:    &single,
		PatternProperties: map[string]*jsonschema.Schema{
			EnvVariablePattern.String(): oneOfStringIntBool,
		},
		AdditionalProperties: jsonschema.FalseSchema,
	}

	props.Set("with", with)

	runProps := jsonschema.NewProperties()
	runProps.Set("run", &jsonschema.Schema{
		Type: "string",
	})
	runProps.Set("uses", not)
	runProps.Set("eval", not)
	oneOfRun := &jsonschema.Schema{
		Required:   []string{"run"},
		Properties: runProps,
	}

	usesProps := jsonschema.NewProperties()
	usesProps.Set("run", not)
	usesProps.Set("eval", not)
	usesProps.Set("uses", &jsonschema.Schema{
		Type: "string",
	})
	oneOfUses := &jsonschema.Schema{
		Required:   []string{"uses"},
		Properties: usesProps,
	}

	evalProps := jsonschema.NewProperties()
	evalProps.Set("run", not)
	evalProps.Set("uses", not)
	evalProps.Set("eval", &jsonschema.Schema{
		Type: "string",
	})
	oneOfEval := &jsonschema.Schema{
		Required:   []string{"eval"},
		Properties: evalProps,
	}

	schema.Properties = props
	schema.OneOf = []*jsonschema.Schema{
		oneOfRun,
		oneOfUses,
		oneOfEval,
	}
}
