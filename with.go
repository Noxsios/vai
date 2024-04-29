package vai

import (
	"fmt"
	"maps"
	"strings"
	"text/template"
)

type WithEntry interface{}

type With map[string]interface{}

func PeformLookups(parent, child, global With, mi MatrixInstance) (With, With, error) {
	logger.Debug("templating", "parent", parent, "child", child, "global", global, "matrix-inst", mi)

	r := make(With)
	
	ng := maps.Clone(global)

	for k, v := range child {
		val := fmt.Sprintf("%s", v)
		fm := template.FuncMap{
			"input": func() string {
				v, ok := parent[k]
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
				ng[k] = val
				return val
			},
		}
		tmpl := template.New("withs").Option("missingkey=error").Delims("${{", "}}")
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

	for k, v := range ng {
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
	return r, ng, nil
}
