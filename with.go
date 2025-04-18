// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package maru2

import (
	"context"
	"fmt"
	"maps"
	"runtime"
	"strings"
	"text/template"

	"github.com/charmbracelet/log"
)

// With is a map of string keys and WithEntry values used to pass parameters to called tasks and within steps
//
// Each key will be mapped to an equivalent environment variable
// when the command is run. eg. `with: {foo: bar}` will be passed
// as `foo=bar` to the command.
type With = map[string]any

func constructTemplateEvaluator(input With, previousOutputs CommandOutputs) *template.Template {
	fm := template.FuncMap{
		"input": func(in string) (any, error) {
			v, ok := input[in]
			if !ok {
				return "", fmt.Errorf("%q does not exist in the map of inputs", in)
			}
			return v, nil
		},
		"from": func(stepName, id string) (any, error) {
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
	return template.New("expression evaluator").Option("missingkey=error").Delims("${{", "}}").Funcs(fm)
}

// TemplateWith templates a With map with the given input and previous outputs
func TemplateWith(ctx context.Context, input, local With, previousOutputs CommandOutputs) (With, error) {
	logger := log.FromContext(ctx)

	if len(local) == 0 {
		return input, nil
	}

	logger.Debug("templating", "input", input, "local", local)

	r := make(With, len(local))

	for k, v := range local {
		val, ok := v.(string)
		// if the val is not a string we can skip templating
		if !ok {
			r[k] = v
			continue
		}
		result, err := TemplateString(input, previousOutputs, val)
		if err != nil {
			return nil, err
		}
		r[k] = result
	}

	logger.Debug("templated", "result", r)

	return r, nil
}

// TemplateString templates a string with the given input and previous outputs
func TemplateString(input With, previousOutputs CommandOutputs, str string) (string, error) {
	tmpl, err := constructTemplateEvaluator(input, previousOutputs).Parse(str)
	if err != nil {
		return "", err
	}
	var result strings.Builder

	if err := tmpl.Execute(&result, struct {
		OS       string
		ARCH     string
		PLATFORM string
	}{
		OS:       runtime.GOOS,
		ARCH:     runtime.GOARCH,
		PLATFORM: fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}); err != nil {
		return "", err
	}
	return result.String(), nil
}

// TemplateWithMap recursively processes a With map and templates all string values
func TemplateWithMap(input With, previousOutputs CommandOutputs, withMap With) (With, error) {
	if withMap == nil {
		return nil, nil
	}

	result := make(With)
	for k, v := range withMap {
		switch val := v.(type) {
		case string:
			templated, err := TemplateString(input, previousOutputs, val)
			if err != nil {
				return nil, err
			}
			result[k] = templated
		case map[string]any:
			nestedMap, err := TemplateWithMap(input, previousOutputs, val)
			if err != nil {
				return nil, err
			}
			result[k] = nestedMap
		case []any:
			templatedSlice, err := templateSlice(input, previousOutputs, val)
			if err != nil {
				return nil, err
			}
			result[k] = templatedSlice
		default:
			result[k] = v
		}
	}
	return result, nil
}

// templateSlice recursively processes a slice and templates all string values
func templateSlice(input With, previousOutputs CommandOutputs, slice []any) ([]any, error) {
	result := make([]any, len(slice))
	for i, v := range slice {
		switch val := v.(type) {
		case string:
			templated, err := TemplateString(input, previousOutputs, val)
			if err != nil {
				return nil, err
			}
			result[i] = templated
		case map[string]any:
			nestedMap, err := TemplateWithMap(input, previousOutputs, val)
			if err != nil {
				return nil, err
			}
			result[i] = nestedMap
		case []any:
			templatedSlice, err := templateSlice(input, previousOutputs, val)
			if err != nil {
				return nil, err
			}
			result[i] = templatedSlice
		default:
			result[i] = v
		}
	}
	return result, nil
}

// MergeWithAndParams merges a With map into an InputMap, handling defaults, logging warnings on deprections, etc...
func MergeWithAndParams(ctx context.Context, with With, params InputMap) (With, error) {
	logger := log.FromContext(ctx)
	merged := maps.Clone(with)

	for name, param := range params {
		if _, ok := merged[name]; !ok {
			if param.Required && merged[name] == nil && param.Default == nil {
				return nil, fmt.Errorf("missing required input: %q", name)
			}
			if merged[name] == nil {
				merged[name] = param.Default
			}
			continue
		}
		// If the input is deprecated AND provided, log a warning
		if param.DeprecatedMessage != "" && with[name] != nil {
			logger.Warnf("input %q is deprecated: %s", name, param.DeprecatedMessage)
		}

		// If the input is provided, and the default is set, ensure the types match
		if param.Default != nil && with[name] != nil {
			switch with[name].(type) {
			case string:
				if _, ok := param.Default.(string); !ok {
					return nil, fmt.Errorf("input %q has type string, but default is %T", name, param.Default)
				}
			case bool:
				if _, ok := param.Default.(bool); !ok {
					return nil, fmt.Errorf("input %q has type bool, but default is %T", name, param.Default)
				}
			case int:
				if _, ok := param.Default.(int); !ok {
					return nil, fmt.Errorf("input %q has type int, but default is %T", name, param.Default)
				}
			default:
				return nil, fmt.Errorf("unknown type for input %q, default is %T, got %T", name, param.Default, with[name])
			}
		}
	}

	return merged, nil
}
