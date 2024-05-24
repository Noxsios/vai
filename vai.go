// Package vai provides a simple task runner.
package vai

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

// Run executes a task in a workflow with the given inputs.
//
// For all `uses` steps, this function will be called recursively.
func Run(ctx context.Context, wf Workflow, taskName string, outer With) error {
	persist := make(With)
	outputs := make(CommandOutputs)

	if taskName == "" {
		taskName = DefaultTaskName
	}

	task, err := wf.Find(taskName)
	if err != nil {
		return err
	}

	for _, step := range task {
		templated, persisted, err := PerformLookups(outer, step.With, persist, outputs)
		if err != nil {
			return err
		}
		for k, v := range persisted {
			persist[k] = v
		}

		switch step.Operation() {
		case OperationUses:
			_, err = wf.Find(step.Uses)
			if err != nil {
				if err := ExecuteUses(ctx, step.Uses, templated); err != nil {
					return err
				}
			} else {
				if err := Run(ctx, wf, step.Uses, templated); err != nil {
					return err
				}
			}
		case OperationRun:
			outFile, err := os.CreateTemp("", "vai-output-*")
			if err != nil {
				return err
			}
			defer os.Remove(outFile.Name())
			defer outFile.Close()

			env := os.Environ()
			for k, v := range templated {
				toEnvVar := func(s string) string {
					return strings.ToUpper(strings.ReplaceAll(s, "-", "_"))
				}

				env = append(env, fmt.Sprintf("%s=%s", toEnvVar(k), v))
			}
			env = append(env, fmt.Sprintf("VAI_OUTPUT=%s", outFile.Name()))
			// TODO: handle other shells
			cmd := exec.CommandContext(ctx, "sh", "-e", "-c", step.Run)
			cmd.Env = env
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin

			customStyles := log.DefaultStyles()
			customStyles.Message = lipgloss.NewStyle().Foreground(lipgloss.Color("#2f333a"))
			logger.SetStyles(customStyles)

			fmt.Println()
			lines := strings.Split(step.Run, "\n")
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
				out, err := ParseOutput(outFile)
				if err != nil {
					return err
				}
				if len(out) == 0 {
					continue
				}
				// TODO: conflicted about whether to save the contents of the file or just the file path
				outputs[step.ID] = out
			}

		default:
			return fmt.Errorf("unknown operation")
		}
	}

	return nil
}
