package vai

import (
	"fmt"
)

type Workflow map[string][]Task

type Matrix map[string][]any

type MatrixInstance map[string]any

func (wf Workflow) Find(call string) ([]Task, error) {
	tasks, ok := wf[call]
	if !ok {
		return nil, fmt.Errorf("task %q not found", call)
	}
	return tasks, nil
}
