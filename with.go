package vai

import (
	"fmt"
	"maps"
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
// 1. Templating: executes the `input`, `default`, `persist`, and `from` functions against the `outer` and `local` With maps
//
// 2. Merging: merges the `persisted` and `local` With maps, with `local` taking precedence
func PerformLookups(input, local, persisted With, previousOutputs CommandOutputs) (With, With, error) {
	logger.Debug("templating", "input", input, "local", local, "persisted", persisted)

	r := make(With)

	persist := maps.Clone(persisted)

	for k, v := range local {
		val := fmt.Sprintf("%s", v)
		fm := template.FuncMap{
			"input": func() string {
				v, ok := input[k]
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
				persist[k] = val
				return val
			},
			"from": func(taskName, name string) string {
				v, ok := previousOutputs[taskName]
				if !ok {
					return ""
				}
				return v[name]
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

	for k, v := range persist {
		_, ok := r[k]
		if ok {
			continue
		}
		r[k] = v
	}

	logger.Debug("templated", "result", r)
	return r, persist, nil
}
