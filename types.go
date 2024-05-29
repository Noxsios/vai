// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"cmp"
	"slices"

	"github.com/invopop/jsonschema"
)

// DefaultTaskName is the default task name
const DefaultTaskName = "default"

// DefaultFileName is the default file name
const DefaultFileName = "vai.yaml"

// Task is a list of steps
type Task []Step

// Workflow is a map of tasks, where the key is the task name
//
// This is the main structure that represents `vai.yaml` and other vai workflow files
type Workflow map[string]Task

// Find returns a task by name
//
// If the task is not found, an error is returned
func (wf Workflow) Find(call string) (Task, bool) {
	task, ok := wf[call]
	return task, ok
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

	schema.ID = "https://raw.githubusercontent.com/Noxsios/vai/main/vai.schema.json"

	schema.AdditionalProperties = jsonschema.FalseSchema

	return schema
}
