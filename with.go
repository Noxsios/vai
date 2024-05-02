package vai

import (
	"fmt"
	"maps"
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

// PerformLookups does the following:
//
// 1. Templating: executes the `input`, `default`, `persist`, and `from` functions against the `outer` and `local` With maps
// 2. Merging: merges the `persisted` and `local` With maps, with `local` taking precedence
// 3. MatrixInstance: merges the `mi` MatrixInstance into the result, with `mi` taking precedence
func PerformLookups(outer, local, persisted With, outputs CommandOutputs, mi MatrixInstance) (With, With, error) {
	logger.Debug("templating", "outer", outer, "local", local, "global", persisted, "matrix-inst", mi)

	r := make(With)

	np := maps.Clone(persisted)

	for k, v := range local {
		val := fmt.Sprintf("%s", v)
		fm := template.FuncMap{
			"input": func() string {
				v, ok := outer[k]
				if !ok || v == "" {
					logger.Warn("no input", "key", k)
					return ""
				}
				return fmt.Sprintf("%s", v)
			},
			"default": func(def, curr string) string {
				if curr == "" {
					return def
				}
				return curr
			},
			"persist": func(val string) string {
				if val == "" {
					return ""
				}
				np[k] = val
				return val
			},
			"from": func(taskName, name string) string {
				v, ok := outputs[taskName]
				if !ok {
					return ""
				}
				return v[name]
			},
		}
		tmpl := template.New("input templating").Option("missingkey=error").Delims("${{", "}}")
		tmpl.Funcs(fm)
		tmpl, err := tmpl.Parse(val)
		if err != nil {
			return nil, nil, err
		}
		var templated strings.Builder
		if err := tmpl.Execute(&templated, nil); err != nil {
			return nil, nil, err
		}
		result := templated.String()
		r[k] = result
	}

	for k, v := range np {
		_, ok := r[k]
		if ok {
			continue
		}
		r[k] = v
	}

	for k, v := range mi {
		_, ok := r[k]
		if ok {
			return nil, nil, fmt.Errorf("matrix key %q already exists in with", k)
		}
		r[k] = v
	}

	logger.Debug("templated", "result", r)
	return r, np, nil
}
