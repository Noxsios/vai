// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"fmt"
	"runtime"
	"strings"
	"text/template"
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
		val := fmt.Sprintf("%s", v)
		fm := template.FuncMap{
			"input": func() string {
				v, ok := input[k]
				if !ok || v == "" {
					return ""
				}
				return fmt.Sprintf("%s", v)
			},
			"default": func(def, curr string) string {
				if len(curr) == 0 {
					return def
				}
				return curr
			},
			"persist": func(s ...string) string {
				toPersist = append(toPersist, k)
				if len(s) == 0 {
					return ""
				}
				return s[0]
			},
			"from": func(stepName, id string) (string, error) {
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
		}
		tmpl := template.New("expression evaluator").Option("missingkey=error").Delims("${{", "}}")
		tmpl.Funcs(fm)
		tmpl, err := tmpl.Parse(val)
		if err != nil {
			return nil, nil, err
		}
		var templated strings.Builder

		if err := tmpl.Execute(&templated, struct {
			OS       string
			ARCH     string
			PLATFORM string
		}{
			OS:       runtime.GOOS,
			ARCH:     runtime.GOARCH,
			PLATFORM: fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		}); err != nil {
			return nil, nil, err
		}
		result := templated.String()
		r[k] = result
	}

	logger.Debug("templated", "result", r)
	return r, toPersist, nil
}
