// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

// Package vai provides a simple task runner.
package vai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/noxsios/vai/modv"
	"github.com/noxsios/vai/uses"
)

// Run executes a task in a workflow with the given inputs.
//
// For all `uses` steps, this function will be called recursively.
func Run(ctx context.Context, store *uses.Store, wf Workflow, taskName string, outer With, origin string) error {
	if taskName == "" {
		taskName = DefaultTaskName
	}

	task, ok := wf.Find(taskName)
	if !ok {
		return fmt.Errorf("task %q not found", taskName)
	}

	outputs := make(CommandOutputs)

	for _, step := range task {
		templated, err := PerformLookups(ctx, outer, step.With, outputs)
		if err != nil {
			return err
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

		if step.Eval != "" {
			printScript(ctx, ">", step.Eval)

			script := tengo.NewScript([]byte(step.Eval))
			mods := stdlib.GetModuleMap(stdlib.AllModuleNames()...)
			mods.AddBuiltinModule("semver", modv.SemverModule)
			script.SetImports(mods)

			for k, v := range templated {
				if err := script.Add(k, v); err != nil {
					return err
				}
			}
			// this addition will not trigger any error conditions from tengo.FromInterface
			_ = script.Add("vai_output", map[string]interface{}{})

			compiled, err := script.Compile()
			if err != nil {
				return err
			}
			if err := compiled.RunContext(ctx); err != nil {
				return err
			}
			if step.ID != "" {
				outputs[step.ID] = compiled.Get("vai_output").Map()
			}
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

				env = append(env, fmt.Sprintf("%s=%s", toEnvVar(k), val))
			}
			env = append(env, fmt.Sprintf("VAI_OUTPUT=%s", outFile.Name()))
			// TODO: handle other shells
			cmd := exec.CommandContext(ctx, "sh", "-e", "-c", step.Run)
			cmd.Env = env
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin

			printScript(ctx, "$", step.Run)

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
				outputs[step.ID] = make(map[string]any)
				for k, v := range out {
					outputs[step.ID][k] = v
				}
			}
		}
	}

	return nil
}

func toEnvVar(s string) string {
	return strings.ToUpper(strings.ReplaceAll(s, "-", "_"))
}
