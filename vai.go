// Package vai provides a simple task runner.
package vai

import (
	"fmt"
	"os"
)

// Force is a global flag to bypass SHA256 checksum verification for cached remote files.
var Force = false

const (
	// CacheEnvVar is the environment variable for the cache directory.
	CacheEnvVar = "VAI_CACHE"
)

// CommandOutputs is a map of step IDs to their outputs.
//
// It is currently NOT goroutine safe.
type CommandOutputs map[string]map[string]string

// Run executes a task in a workflow with the given inputs.
//
// For all steps that have a `uses` step, this function will be called recursively.
func Run(wf Workflow, taskName string, outer With) error {
	global := make(With)
	outputs := make(CommandOutputs)

	task, err := wf.Find(taskName)
	if err != nil {
		return err
	}

	for _, step := range task {
		instances := make([]MatrixInstance, 0)
		for k, v := range step.Matrix {
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
		if step.Uses != nil && step.CMD != nil {
			return fmt.Errorf("step cannot have both cmd and uses")
		}
		for _, mi := range instances {
			w, ng, err := PeformLookups(outer, step.With, global, outputs, mi)
			if err != nil {
				return err
			}
			for k, v := range ng {
				global[k] = v
			}

			switch step.Operation() {
			case OperationUses:
				_, err := wf.Find(step.Uses.String())
				if err != nil {
					if err := step.Uses.Run(w); err != nil {
						return err
					}
				} else {
					if err := Run(wf, step.Uses.String(), w); err != nil {
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
				if err := step.Run(w, outFile); err != nil {
					return err
				}

				if step.ID != "" {
					fi, err := outFile.Stat()
					if err != nil {
						return err
					}

					if fi.Size() > 0 {
						outputs[step.ID] = make(map[string]string)
						out, err := ParseOutputFile(outFile)
						if err != nil {
							return err
						}
						outputs[step.ID] = out
					}
				}

			default:
				return fmt.Errorf("unknown operation")
			}
		}
	}

	return nil
}
