// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

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
	"github.com/noxsios/vai/storage"
)

// Run executes a task in a workflow with the given inputs.
//
// For all `uses` steps, this function will be called recursively.
func Run(ctx context.Context, store *storage.Store, wf Workflow, taskName string, outer With, origin string) error {
	if taskName == "" {
		taskName = DefaultTaskName
	}

	task, ok := wf.Find(taskName)
	if !ok {
		return fmt.Errorf("task %q not found", taskName)
	}

	persist := make(With)
	outputs := make(CommandOutputs)

	for _, step := range task {
		templated, toPersist, err := PerformLookups(outer, step.With, outputs)
		if err != nil {
			return err
		}
		for _, k := range toPersist {
			persist[k] = templated[k]
		}

		if step.Uses != "" {
			if _, ok := wf.Find(step.Uses); ok {
				if err := Run(ctx, store, wf, step.Uses, templated, origin); err != nil {
					return err
				}
				continue
			}
			if err := ExecuteUses(ctx, store, step.Uses, templated, origin); err != nil {
				return err
			}
			continue
		}

		if step.Run != "" {
			outFile, err := os.CreateTemp("", "vai-output-*")
			if err != nil {
				return err
			}
			defer os.Remove(outFile.Name())
			defer outFile.Close()

			env := os.Environ()
			for k, v := range templated {
				var val string
				switch v := v.(type) {
				case string:
					val = v
				case fmt.Stringer:
					val = v.String()
				case int:
					val = fmt.Sprintf("%d", v)
				case bool:
					val = fmt.Sprintf("%t", v)
				default:
					val = fmt.Sprintf("%v", v)
				}

				env = append(env, fmt.Sprintf("%s=%s", toEnvVar(k), val))
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

			lines := strings.Split(step.Run, "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" {
					continue
				}
				logger.Printf("$ %s", trimmed)
			}

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
		}
	}

	return nil
}

func toEnvVar(s string) string {
	return strings.ToUpper(strings.ReplaceAll(s, "-", "_"))
}
