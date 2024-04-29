package vai

import (
	"fmt"
)

func Run(tasks []Task, outer With) error {
	global := make(With)
	for _, t := range tasks {
		instances := make([]MatrixInstance, 0)
		for k, v := range t.Matrix {
			for _, i := range v {
				mi := make(MatrixInstance)
				mi[k] = i
				instances = append(instances, mi)
			}
		}
		if len(instances) == 0 {
			instances = append(instances, MatrixInstance{})
		}
		if t.Uses != nil && t.CMD != nil {
			return fmt.Errorf("task cannot have both cmd and uses")
		}
		for _, mi := range instances {
			w, ng, err := PeformLookups(outer, t.With, global, mi)
			if err != nil {
				return err
			}
			for k, v := range ng {
				global[k] = v
			}

			switch t.Operation() {
			case OperationUses:
				if err := t.Uses.Run(w); err != nil {
					return err
				}
			case OperationRun:
				if err := t.Run(w); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unknown operation")
			}
		}
	}

	return nil
}
