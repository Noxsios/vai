package vai

import (
	"fmt"
)

type Task []Step

type Workflow map[string]Task

// TODO: schema validation
type Matrix map[string][]any

type MatrixInstance map[string]any

func (wf Workflow) Find(call string) (Task, error) {
	task, ok := wf[call]
	if !ok {
		return nil, fmt.Errorf("task %q not found", call)
	}
	return task, nil
}
