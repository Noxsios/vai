// Package vai provides a simple task runner.
package vai

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
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

	if taskName == "" {
		taskName = DefaultTaskName
	}

	task, err := wf.Find(taskName)
	if err != nil {
		return err
	}

	for _, step := range task {
		templated, persisted, err := PerformLookups(outer, step.With, global, outputs)
		if err != nil {
			return err
		}
		for k, v := range persisted {
			global[k] = v
		}

		switch step.Operation() {
		case OperationUses:
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
			for _, mi := range instances {
				templated, err := templated.MergeMatrixInstance(mi)
				if err != nil {
					return err
				}
				_, err = wf.Find(step.Uses)
				if err != nil {
					if err := ExecuteUses(step.Uses, templated); err != nil {
						return err
					}
				} else {
					if err := Run(wf, step.Uses, templated); err != nil {
						return err
					}
				}
			}
		case OperationRun:
			outFile, err := os.CreateTemp("", "vai-output-*")
			if err != nil {
				return err
			}
			outFile.Close()
			defer os.Remove(outFile.Name())

			env := os.Environ()
			for k, v := range templated {
				toEnvVar := func(s string) string {
					return strings.ToUpper(strings.ReplaceAll(s, "-", "_"))
				}

				env = append(env, fmt.Sprintf("%s=%s", toEnvVar(k), v))
			}
			env = append(env, fmt.Sprintf("VAI_OUTPUT=%s", outFile.Name()))
			// TODO: handle other shells
			cmd := exec.Command("sh", "-e", "-c", step.CMD)
			cmd.Env = env
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin

			customStyles := log.DefaultStyles()
			customStyles.Message = lipgloss.NewStyle().Foreground(lipgloss.Color("#2f333a"))
			logger.SetStyles(customStyles)

			fmt.Println()
			lines := strings.Split(step.CMD, "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" {
					continue
				}
				logger.Printf("$ %s", trimmed)
			}
			fmt.Println()
			logger.SetStyles(log.DefaultStyles())

			if err := cmd.Run(); err != nil {
				return err
			}

			if step.ID != "" {
				outFile, err := os.Open(outFile.Name())
				if err != nil {
					return err
				}

				fi, err := outFile.Stat()
				if err != nil {
					return err
				}

				if fi.Size() > 0 {
					outputs[step.ID] = make(map[string]string)
					out, err := ParseOutputFile(outFile.Name())
					if err != nil {
						return err
					}
					// TODO: conflicted about whether to save the contents of the file or the file path
					outputs[step.ID] = out
				}
			}

		default:
			return fmt.Errorf("unknown operation")
		}
	}

	return nil
}

// ParseOutputFile parses the output file of a step
func ParseOutputFile(loc string) (map[string]string, error) {
	f, err := os.Open(loc)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	// error if larger than 50MB, same limits as GitHub Actions
	if fi.Size() > 50*1024*1024 {
		return nil, fmt.Errorf("output file too large")
	}

	scanner := bufio.NewScanner(f)
	result := make(map[string]string)
	var currentKey string
	var currentDelimiter string
	var multiLineValue []string
	var collecting bool

	for scanner.Scan() {
		line := scanner.Text()

		if collecting {
			if line == currentDelimiter {
				// End of multiline value
				value := strings.Join(multiLineValue, "\n")
				result[currentKey] = value
				collecting = false
				multiLineValue = []string{}
				currentKey = ""
				currentDelimiter = ""
			} else {
				multiLineValue = append(multiLineValue, line)
			}
			continue
		}

		if idx := strings.Index(line, "="); idx != -1 {
			// Split the line at the first '=' to handle the key-value pair
			key := line[:idx]
			value := line[idx+1:]
			// Check if the value is a potential start of a multiline value
			if strings.HasSuffix(value, "<<") {
				currentKey = key
				currentDelimiter = strings.TrimSpace(value[2:])
				collecting = true
			} else {
				result[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Handle case where file ends but multiline was being collected
	if collecting && len(multiLineValue) > 0 {
		value := strings.Join(multiLineValue, "\n")
		result[currentKey] = value
	}

	return result, nil
}
