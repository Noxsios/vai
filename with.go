// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"context"
	"fmt"
	"runtime"

	"github.com/charmbracelet/log"
	"github.com/d5/tengo/v2"
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
func PerformLookups(ctx context.Context, outer, local With, previousOutputs CommandOutputs) (With, error) {
	if len(local) == 0 {
		return local, nil
	}

	logger := log.FromContext(ctx)

	logger.Debug("templating", "input", outer, "local", local)

	r := make(With, len(local))

	for k, v := range local {
		val, ok := v.(string)
		if !ok {
			r[k] = v
			continue
		}

		env := map[string]interface{}{
			"os":       runtime.GOOS,
			"arch":     runtime.GOARCH,
			"platform": fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			"input":    outer[k],
		}

		steps := map[string]tengo.Object{}

		for k, v := range previousOutputs {
			obj, err := tengo.FromInterface(v)
			if err != nil {
				return nil, err
			}
			steps[k] = obj
		}

		env["steps"] = steps

		out, err := tengo.Eval(ctx, val, env)
		if err != nil {
			return nil, err
		}
		if out == nil {
			return nil, fmt.Errorf("expression evaluated to <nil>:\n\t%s", val)
		}
		r[k] = out
	}

	logger.Debug("templated", "result", r)
	return r, nil
}
