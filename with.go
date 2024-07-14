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

// PerformLookups evaluates the expressions in the local With map
func PerformLookups(outer, local With, previousOutputs CommandOutputs) (With, error) {
	if len(local) == 0 {
		return local, nil
	}

	logger.Debug("templating", "input", outer, "local", local)

	r := make(With, len(local))

	for k, v := range local {
		env := map[string]interface{}{
			"os":       runtime.GOOS,
			"arch":     runtime.GOARCH,
			"platform": fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			"input":    outer[k],
		}

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
		case int, bool, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			// no need to evaluate
			r[k] = v
			continue
		default:
			// should never happen in CLI due to schema validation
			return nil, fmt.Errorf("unsupported type %T for key %q", v, k)
		}

		program, err := expr.Compile(script, expr.Env(env), from)
		if err != nil {
			return nil, err
		}

		out, err := expr.Run(program, env)
		if err != nil {
			return nil, err
		}
		// ensure output is not nil
		if out == nil {
			return nil, fmt.Errorf("expression %q evaluated to <nil>", script)
		}
		r[k] = out
	}

	logger.Debug("templated", "result", r)
	return r, nil
}
