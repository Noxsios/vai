package vai

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/invopop/jsonschema"
)

// DefaultTaskName is the default task name
const DefaultTaskName = "default"

// DefaultFileName is the default file name
const DefaultFileName = "vai.yaml"

// Task is a list of steps
type Task []Step

// Matrix is a map[]{string|int|bool}
//
// Type safety cannot currently be enforced at compile time,
// and is instead enforced at runtime using JSON schema validation
//
// example (YAML):
//
//	matrix:
//	  os: [linux, darwin]
//	  arch: [amd64, arm64]
type Matrix map[string][]any

// MatrixInstance is a map[string]{string|int|bool}
//
// Type safety cannot currently be enforced at compile time,
// and is instead enforced at runtime using JSON schema validation
//
// example:
//
//	mi := MatrixInstance{
//	  "os": "linux",
//	  "latest": true,
//	}
type MatrixInstance map[string]any

// Workflow is a map of tasks, where the key is the task name
//
// This is the main structure that represents `vai.yaml` and other vai workflow files
type Workflow map[string]Task

// Find returns a task by name
//
// If the task is not found, an error is returned
func (wf Workflow) Find(call string) (Task, error) {
	task, ok := wf[call]
	if !ok {
		return nil, fmt.Errorf("task %q not found", call)
	}
	return task, nil
}

// OrderedTaskNames returns a list of task names in alphabetical order
//
// The default task is always first
func (wf Workflow) OrderedTaskNames() []string {
	names := make([]string, 0, len(wf))
	for k := range wf {
		names = append(names, k)
	}
	slices.SortStableFunc(names, func(a, b string) int {
		if a == DefaultTaskName {
			return -1
		}
		if b == DefaultTaskName {
			return 1
		}
		return cmp.Compare(a, b)
	})
	return names
}

// WorkFlowSchema returns a JSON schema for a vai workflow
func WorkFlowSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{}
	reflector.ExpandedStruct = true
	schema := reflector.Reflect(&Workflow{})

	schema.PatternProperties = map[string]*jsonschema.Schema{
		TaskNamePattern.String(): {
			Ref:         "#/$defs/Task",
			Description: "Name of the task",
		},
	}

	schema.AdditionalProperties = jsonschema.FalseSchema

	return schema
}
