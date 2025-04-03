// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package maru2

import (
	"slices"

	"github.com/invopop/jsonschema"
)

// InputMap is a map of input parameters for a workflos
type InputMap map[string]InputParameter

// InputParameter represents a single input parameter for a task, to be used w/ `with`
type InputParameter struct {
	Description       string `json:"description" jsonschema:"description=Description of the parameter,required"`
	DeprecatedMessage string `json:"deprecatedMessage,omitempty" jsonschema:"description=Message to display when the parameter is deprecated"`
	Required          bool   `json:"required,omitempty" jsonschema:"description=Whether the parameter is required,default=true"`
	Default           any    `json:"default,omitempty" jsonschema:"description=Default value for the parameter"`
}

// JSONSchemaExtend extends the JSON schema for a step
func (InputParameter) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.Properties.Set("default", &jsonschema.Schema{
		Description: "Default value for the parameter",
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
	})
}

// Step is a single step in a task
//
// While a step can have any combination of `run`, and `uses` fields, only one of them should be set
// at a time.
//
// This is enforced by JSON schema validation.
type Step struct {
	// Run is the command/script to run
	Run string `json:"run,omitempty"`
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
		Examples:    []interface{}{"echo 'Hello World'", "cat file.txt | grep pattern"},
	})
	props.Set("uses", &jsonschema.Schema{
		Type:        "string",
		Description: "Location of a remote task to call conforming to the purl spec",
		Examples:    []interface{}{"builtin:echo", "pkg:github/defenseunicorns/maru2@main?task=echo"},
	})
	props.Set("id", &jsonschema.Schema{
		Type:        "string",
		Description: "Unique identifier for the step",
		Examples:    []interface{}{"setup", "build", "test"},
	})
	props.Set("name", &jsonschema.Schema{
		Type:        "string",
		Description: "Human-readable name for the step",
		Examples:    []interface{}{"Setup environment", "Build application", "Run tests"},
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
	oneOfRun := &jsonschema.Schema{
		Required:   []string{"run"},
		Properties: runProps,
	}

	usesProps := jsonschema.NewProperties()
	usesProps.Set("run", not)
	usesProps.Set("uses", &jsonschema.Schema{
		Type: "string",
	})
	oneOfUses := &jsonschema.Schema{
		Required:   []string{"uses"},
		Properties: usesProps,
	}

	var allBuiltinSchemas []*jsonschema.Schema

	for name, builtin := range Builtins {
		builtinSchema := &jsonschema.Schema{
			If: &jsonschema.Schema{
				Properties: jsonschema.NewProperties(),
			},
			Then: &jsonschema.Schema{
				Properties: jsonschema.NewProperties(),
			},
		}

		builtinSchema.If.Properties.Set("uses", &jsonschema.Schema{
			Type:    "string",
			Pattern: "^builtin:" + name + "(@.*)?$",
		})

		withSchema := &jsonschema.Schema{
			Type:                 "object",
			AdditionalProperties: jsonschema.FalseSchema,
			Properties:           jsonschema.NewProperties(),
		}

		var required []string

		var names []string
		for name := range builtin.Params {
			names = append(names, name)
		}
		slices.Sort(names)
		for _, paramName := range names {
			param := builtin.Params[paramName]
			paramSchema := &jsonschema.Schema{
				Description: param.Description,
			}

			if param.Default != nil {
				switch v := param.Default.(type) {
				case string:
					paramSchema.Type = "string"
					paramSchema.Default = v
				case int:
					paramSchema.Type = "integer"
					paramSchema.Default = v
				case bool:
					paramSchema.Type = "boolean"
					paramSchema.Default = v
				default:
					paramSchema = oneOfStringIntBool
				}
			} else {
				paramSchema = oneOfStringIntBool
			}

			withSchema.Properties.Set(paramName, paramSchema)

			if param.Required {
				required = append(required, paramName)
			}
		}

		withSchema.Required = required

		thenProps := jsonschema.NewProperties()
		thenProps.Set("with", withSchema)
		builtinSchema.Then.Properties = thenProps

		if len(required) > 0 {
			builtinSchema.Then.Required = []string{"with"}
		}

		allBuiltinSchemas = append(allBuiltinSchemas, builtinSchema)
	}

	oneOfUses.AllOf = allBuiltinSchemas

	schema.Properties = props
	schema.OneOf = []*jsonschema.Schema{
		oneOfRun,
		oneOfUses,
	}
}
