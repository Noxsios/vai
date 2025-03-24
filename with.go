// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package maru2

import (
	"context"
	"fmt"
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
type With map[string]any

func constructTemplateEvaluator(input With, previousOutputs CommandOutputs) *template.Template {
	fm := template.FuncMap{
		"input": func(in string) (any, error) {
			v, ok := input[in]
			if !ok {
				return "", fmt.Errorf("%q does not exist in the map of inputs", in)
			}
			return v, nil
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
	return template.New("expression evaluator").Option("missingkey=error").Delims("${{", "}}").Funcs(fm)
}

func TemplateWith(ctx context.Context, input, local With, previousOutputs CommandOutputs) (With, error) {
	logger := log.FromContext(ctx)
	logger.Debug("templating", "input", input, "local", local)

	if len(local) == 0 {
		return input, nil
	}

	r := make(With, len(local))

	for k, v := range local {
		val, ok := v.(string)
		// if the val is not a string we can skip templating
		if !ok {
			r[k] = v
			continue
		}
		tmpl, err := constructTemplateEvaluator(input, previousOutputs).Parse(val)
		if err != nil {
			return nil, err
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
			return nil, err
		}
		r[k] = result.String()
	}

	logger.Debug("templated", "result", r)

	return r, nil
}

// TemplateRun
func TemplateRun(run string, input With, previousOutputs CommandOutputs) (string, error) {
	tmpl, err := constructTemplateEvaluator(input, previousOutputs).Parse(run)
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
