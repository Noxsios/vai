package vai

import (
	"fmt"
)

type Workflow map[string][]Task

type Matrix map[string][]interface{}

type MatrixInstance map[string]interface{}

func (wf Workflow) Find(call string) ([]Task, error) {
	logger.Debug("finding", "task", call)
	tasks, ok := wf[call]
	if !ok {
		return nil, fmt.Errorf("task %q not found", call)
	}
	return tasks, nil
}
