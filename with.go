package vai

import (
	"fmt"
	"strings"
	"text/template"
)

type With map[string]interface{}

func PeformLookups(parent, child With, mi MatrixInstance) (With, error) {
	logger.Debug("templating", "parent", parent, "child", child, "instance", mi)

	r := make(With)

	for k, v := range child {
		val := fmt.Sprintf("%s", v)
		fm := template.FuncMap{
			"input": func(fallback ...any) any {
				v, ok := parent[k]
				if !ok || v == "" {
					if len(fallback) == 0 {
						logger.Warn("no input", "key", k)
						return nil
					}
					return fallback[0]
				}
				return v
			},
		}
		tmpl := template.New("withs").Option("missingkey=error").Delims("${{", "}}")
		tmpl.Funcs(fm)
		tmpl, err := tmpl.Parse(val)
		if err != nil {
			return nil, err
		}
		var templated strings.Builder
		if err := tmpl.Execute(&templated, nil); err != nil {
			return nil, err
		}
		result := templated.String()
		r[k] = result
	}

	for k, v := range mi {
		_, ok := r[k]
		if !ok {
			r[k] = v
		} else {
			return nil, fmt.Errorf("matrix key %q already exists in with", k)
		}
	}

	logger.Debug("templated", "result", r)
	return r, nil
}