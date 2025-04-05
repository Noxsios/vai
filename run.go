// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

// Package maru2 provides a simple task runner.
package maru2

import (
	"context"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/log"
)

// Run executes a task in a workflow with the given inputs.
//
// For all `uses` steps, this function will be called recursively.
func Run(ctx context.Context, wf Workflow, taskName string, outer With, origin string, dry bool) error {
	if taskName == "" {
		taskName = DefaultTaskName
	}

	task, ok := wf.Tasks.Find(taskName)
	if !ok {
		return fmt.Errorf("task %q not found", taskName)
	}

	outputs := make(CommandOutputs)

	withDefaults, err := MergeWithAndParams(ctx, outer, wf.Inputs)
	if err != nil {
		return err
	}

	var firstError error
	logger := log.FromContext(ctx)

	for _, step := range task {
		if firstError == nil && step.If == "failure" {
			logger.Debug("skipping step", "name", step.Name, "if", step.If)
			continue
		}

		if firstError != nil && step.If == "" {
			logger.Debug("skipping step", "name", step.Name, "if", step.If)
			continue
		}

		if step.Uses != "" {
			templatedWith, err := TemplateWith(ctx, withDefaults, step.With, outputs)
			if err != nil {
				return err
			}
			if _, ok := wf.Tasks.Find(step.Uses); ok {
				if err := Run(ctx, wf, step.Uses, templatedWith, origin, dry); err != nil {
					if firstError == nil { // subsequent errors are ignored
						firstError = err
					}
				}
				continue
			}
			if err := ExecuteUses(ctx, step.Uses, templatedWith, origin, dry); err != nil {
				if firstError == nil { // subsequent errors are ignored
					firstError = err
				}
			}
			continue
		}

		if step.Run != "" {
			templatedRun, err := TemplateString(withDefaults, outputs, step.Run)
			if err != nil {
				return err
			}

			printScript(ctx, "$", templatedRun)
			if dry {
				continue
			}

			outFile, err := os.CreateTemp("", "maru2-output-*")
			if err != nil {
				if firstError == nil { // subsequent errors are ignored
					firstError = err
				}
				continue
			}
			defer os.Remove(outFile.Name())
			defer outFile.Close()

			env := os.Environ()
			// TODO: not a big fan of this
			for k, v := range withDefaults {
				var val string
				switch v := v.(type) {
				case string:
					val = v
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
					val = fmt.Sprintf("%d", v)
				case bool:
					val = fmt.Sprintf("%t", v)
					// todo: what about the default case?
					// through schema validation we know that the value is a string|int|bool
				}

				env = append(env, fmt.Sprintf("INPUT_%s=%s", toEnvVar(k), val))
			}
			env = append(env, fmt.Sprintf("MARU2_OUTPUT=%s", outFile.Name()))
			// TODO: handle other shells
			cmd := exec.CommandContext(ctx, "sh", "-e", "-c", templatedRun)
			cmd.Env = env
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin

			if err := cmd.Run(); err != nil {
				if firstError == nil { // subsequent errors are ignored
					firstError = err
				}
				continue
			}

			if step.ID != "" {
				out, err := ParseOutput(outFile)
				if err != nil {
					if firstError == nil { // subsequent errors are ignored
						firstError = err
					}
					continue
				}
				if len(out) == 0 {
					continue
				}
				outputs[step.ID] = make(map[string]string)
				maps.Copy(outputs[step.ID], out)
			}
		}
	}

	return firstError
}

func toEnvVar(s string) string {
	return strings.ToUpper(strings.ReplaceAll(s, "-", "_"))
}
