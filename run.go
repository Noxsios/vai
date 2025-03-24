// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

// Package maru2 provides a simple task runner.
package maru2

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"strings"
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

	withDefaults := outer
	for k, v := range wf.Inputs {
		// TODO: actually think of a strategy
		name := strings.TrimLeft(strings.ToLower(k), "$")
		if v.Required && withDefaults[name] == nil && v.Default == nil {
			return fmt.Errorf("missing required input: %s", k)
		}
		if withDefaults[name] == nil {
			withDefaults[name] = v.Default
		}
	}
	// --with + defaults

	for _, step := range task {
		if step.Uses != "" {
			templatedWith, err := TemplateWith(ctx, withDefaults, step.With, outputs)
			if err != nil {
				return err
			}
			if _, ok := wf.Tasks.Find(step.Uses); ok {
				if err := Run(ctx, wf, step.Uses, templatedWith, origin, dry); err != nil {
					return err
				}
				continue
			}
			if err := ExecuteUses(ctx, step.Uses, templatedWith, origin, dry); err != nil {
				return err
			}
			continue
		}

		if step.Run != "" {
			templated := withDefaults

			templatedRun, err := TemplateRun(step.Run, templated, outputs)
			if err != nil {
				return err
			}

			printScript(ctx, "$", templatedRun)
			if dry {
				continue
			}

			outFile, err := os.CreateTemp("", "maru2-output-*")
			if err != nil {
				return err
			}
			defer os.Remove(outFile.Name())
			defer outFile.Close()

			env := os.Environ()
			// TODO: not a big fan of this
			for k, v := range templated {
				var val string
				switch v := v.(type) {
				case string:
					val = v
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
					val = fmt.Sprintf("%d", v)
				case bool:
					val = fmt.Sprintf("%t", v)
				default:
					// JSON marshal all other types
					b, err := json.Marshal(v)
					if err != nil {
						return err
					}
					val = string(b)
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
				outputs[step.ID] = make(map[string]string)
				maps.Copy(outputs[step.ID], out)
			}
		}
	}

	return nil
}

func toEnvVar(s string) string {
	return strings.ToUpper(strings.ReplaceAll(s, "-", "_"))
}
