// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"fmt"
	"runtime"

	"github.com/expr-lang/expr"
)

// WithEntry is a single entry in a With map
type WithEntry any

// With is a map of string keys and WithEntry values used to pass parameters to called tasks and within steps
//
// Each key will be mapped to an equivalent environment variable
// when the command is run. eg. `with: {foo: bar}` will be passed
// as `foo=bar` to the command.
type With map[string]WithEntry

// PerformLookups performs the following:
//
// 1. Templating: executes the `input`, `default`, `persist`, and `from` functions against the `input` and `local` With maps
//
// 2. Merging: merges the `persisted` and `local` With maps, with `local` taking precedence
func PerformLookups(input, local With, previousOutputs CommandOutputs) (With, []string, error) {
	if len(local) == 0 {
		return local, nil, nil
	}

	logger.Debug("templating", "input", input, "local", local)

	r := make(With, len(local))
	toPersist := make([]string, 0, len(local))

	for k, v := range local {
		env := map[string]interface{}{
			"os":       runtime.GOOS,
			"arch":     runtime.GOARCH,
			"platform": fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			"input":    input[k],
		}

		persist := expr.Function("persist", func(params ...any) (any, error) {
			if len(params) == 0 || params[0] == nil {
				return nil, fmt.Errorf("no value to persist")
			}
			toPersist = append(toPersist, k)
			return params[0], nil
		})

		from := expr.Function("from", func(params ...any) (any, error) {
			stepName := params[0].(string)
			id := params[1].(string)

			stepOutputs, ok := previousOutputs[stepName]
			if !ok {
				return "", fmt.Errorf("no outputs for step %q", stepName)
			}

			v, ok := stepOutputs[id]
			if ok {
				return v, nil
			}
			return "", fmt.Errorf("no output %q from %q", id, stepName)
		},
			new(func(string, string) (string, error)),
		)

		var script string

		switch v := v.(type) {
		case string:
			script = v
		case int:
		case bool:
			// no need to evaluate
			r[k] = v
			continue
		default:
			// should never happen in CLI due to schema validation
			return nil, nil, fmt.Errorf("unsupported type %T for key %q", v, k)
		}

		program, err := expr.Compile(script, expr.Env(env), persist, from)
		if err != nil {
			return nil, nil, err
		}

		out, err := expr.Run(program, env)
		if err != nil {
			return nil, nil, err
		}
		// ensure output is not nil
		if out == nil {
			return nil, nil, fmt.Errorf("expression %q evaluated to <nil>", script)
		}
		r[k] = out
	}

	logger.Debug("templated", "result", r)
	return r, toPersist, nil
}
