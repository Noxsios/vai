package vai

import (
	"fmt"
	"os"
)

type CommandOutputs map[string]map[string]string

func Run(wf Workflow, taskName string, outer With) error {
	global := make(With)
	outputs := make(CommandOutputs)

	tasks, err := wf.Find(taskName)
	if err != nil {
		return err
	}

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
		// TODO: this will be handled by schema validation on read
		if t.Uses != nil && t.CMD != nil {
			return fmt.Errorf("task cannot have both cmd and uses")
		}
		for _, mi := range instances {
			w, ng, err := PeformLookups(outer, t.With, global, outputs, mi)
			if err != nil {
				return err
			}
			for k, v := range ng {
				global[k] = v
			}

			switch t.Operation() {
			case OperationUses:
				_, err := wf.Find(t.Uses.String())
				if err != nil {
					if err := t.Uses.Run(w); err != nil {
						return err
					}
				} else {
					if err := Run(wf, t.Uses.String(), w); err != nil {
						return err
					}
				}
			case OperationRun:
				outFile, err := os.CreateTemp("", "vai-output-")
				if err != nil {
					return err
				}
				defer func() {
					outFile.Close()
					os.Remove(outFile.Name())
				}()
				if err := t.Run(w, outFile); err != nil {
					return err
				}

				if t.ID != "" {
					fi, err := outFile.Stat()
					if err != nil {
						return err
					}

					if fi.Size() > 0 {
						outputs[t.ID] = make(map[string]string)
						out, err := ParseOutputFile(outFile)
						if err != nil {
							return err
						}
						outputs[t.ID] = out
					}
				}

			default:
				return fmt.Errorf("unknown operation")
			}
		}
	}

	return nil
}
