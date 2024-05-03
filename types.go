package vai

import (
	"fmt"
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

// Matrix is a map[]{string|int|bool}
//
// Type safety cannot currently be enforced at compile time,
// and is instead enforced at runtime using JSON schema validation
//
// example (YAML):
//   matrix:
//     os: [linux, darwin]
//     arch: [amd64, arm64]
type Matrix map[string][]any

// MatrixInstance is a map[string]{string|int|bool}
//
// Type safety cannot currently be enforced at compile time,
// and is instead enforced at runtime using JSON schema validation
//
// example:
//   mi := MatrixInstance{
//     "os": "linux",
//     "latest": true,
//   }
type MatrixInstance map[string]any

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
