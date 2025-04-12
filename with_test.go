// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package maru2

import (
	"context"
	"runtime"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateString(t *testing.T) {
	tests := []struct {
		name           string
		input          With
		previousOutput CommandOutputs
		str            string
		expectedResult string
		expectedError  string
	}{
		{
			name:           "no template",
			str:            "hello world",
			expectedResult: "hello world",
		},
		{
			name:           "with input",
			input:          With{"name": "test"},
			str:            "hello ${{ input \"name\" }}",
			expectedResult: "hello test",
		},
		{
			name:          "with missing input",
			input:         With{},
			str:           "hello ${{ input \"name\" }}",
			expectedError: "\"name\" does not exist in the map of inputs",
		},
		{
			name: "with previous output",
			previousOutput: CommandOutputs{
				"step1": map[string]any{
					"result": "success",
				},
			},
			str:            "status: ${{ from \"step1\" \"result\" }}",
			expectedResult: "status: success",
		},
		{
			name:           "with missing previous output",
			previousOutput: CommandOutputs{},
			str:            "status: ${{ from \"step1\" \"result\" }}",
			expectedError:  "no outputs for step \"step1\"",
		},
		{
			name:           "with OS variable",
			str:            "OS: ${{ .OS }}",
			expectedResult: "OS: " + runtime.GOOS,
		},
		{
			name:           "with ARCH variable",
			str:            "ARCH: ${{ .ARCH }}",
			expectedResult: "ARCH: " + runtime.GOARCH,
		},
		{
			name:           "with PLATFORM variable",
			str:            "PLATFORM: ${{ .PLATFORM }}",
			expectedResult: "PLATFORM: " + runtime.GOOS + "/" + runtime.GOARCH,
		},
		{
			name:  "with multiple variables",
			input: With{"name": "test"},
			previousOutput: CommandOutputs{
				"step1": map[string]any{
					"result": "success",
				},
			},
			str:            "Hello ${{ input \"name\" }}, status: ${{ from \"step1\" \"result\" }}, OS: ${{ .OS }}",
			expectedResult: "Hello test, status: success, OS: " + runtime.GOOS,
		},
		{
			name:          "invalid template syntax",
			str:           "Hello ${{ input",
			expectedError: "unclosed action",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := TemplateString(tc.input, tc.previousOutput, tc.str)

			if tc.expectedError == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}

func TestMergeWithAndParams(t *testing.T) {
	tests := []struct {
		name          string
		with          With
		params        InputMap
		expectedWith  With
		expectedError string
	}{
		{
			name:         "empty inputs",
			with:         With{},
			params:       InputMap{},
			expectedWith: With{},
		},
		{
			name: "with default values",
			with: With{},
			params: InputMap{
				"name": InputParameter{
					Description: "Name parameter",
					Default:     "default-name",
				},
				"version": InputParameter{
					Description: "Version parameter",
					Default:     "1.0.0",
				},
			},
			expectedWith: With{
				"name":    "default-name",
				"version": "1.0.0",
			},
		},
		{
			name: "with overridden values",
			with: With{
				"name": "custom-name",
			},
			params: InputMap{
				"name": InputParameter{
					Description: "Name parameter",
					Default:     "default-name",
				},
				"version": InputParameter{
					Description: "Version parameter",
					Default:     "1.0.0",
				},
			},
			expectedWith: With{
				"name":    "custom-name",
				"version": "1.0.0",
			},
		},
		{
			name: "with required parameter missing",
			with: With{},
			params: InputMap{
				"name": InputParameter{
					Description: "Name parameter",
					Required:    true,
				},
			},
			expectedError: "missing required input: \"name\"",
		},
		{
			name: "with required parameter provided",
			with: With{
				"name": "custom-name",
			},
			params: InputMap{
				"name": InputParameter{
					Description: "Name parameter",
					Required:    true,
				},
			},
			expectedWith: With{
				"name": "custom-name",
			},
		},
		{
			name: "with deprecated parameter",
			with: With{
				"old-param": "value",
			},
			params: InputMap{
				"old-param": InputParameter{
					Description:       "Old parameter",
					DeprecatedMessage: "Use new-param instead",
				},
			},
			expectedWith: With{
				"old-param": "value",
			},
		},
		{
			name: "with extra parameters",
			with: With{
				"name":    "custom-name",
				"extra":   "extra-value",
				"another": 123,
			},
			params: InputMap{
				"name": InputParameter{
					Description: "Name parameter",
					Default:     "default-name",
				},
			},
			expectedWith: With{
				"name":    "custom-name",
				"extra":   "extra-value",
				"another": 123,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			logger := log.New(nil) // Use nil writer for tests
			ctx = log.WithContext(ctx, logger)

			result, err := MergeWithAndParams(ctx, tc.with, tc.params)

			if tc.expectedError == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedWith, result)
			} else {
				require.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestPerformLookups(t *testing.T) {
	testCases := []struct {
		name              string
		input             With
		local             With
		previous          CommandOutputs
		expectedTemplated With
		expectedError     string
	}{
		{
			name: "no lookups",
		},
		{
			name: "invalid template",
			local: With{
				"foo": `${{ input`,
			},
			expectedError: "template: expression evaluator:1: unclosed action",
		},
		{
			name: "simple lookup + builtins",
			input: With{
				"key": "value",
			},
			local: With{
				"key":      "${{ input \"key\" }}",
				"os":       "${{ .OS }}",
				"arch":     "${{ .ARCH }}",
				"platform": "${{ .PLATFORM }}",
				"int":      1,
				"bool":     false,
			},
			expectedTemplated: With{
				"key":      "value",
				"os":       runtime.GOOS,
				"arch":     runtime.GOARCH,
				"platform": runtime.GOOS + "/" + runtime.GOARCH,
				"int":      1,
				"bool":     false,
			},
		},
		{
			name: "missing input",
			local: With{
				"key": `${{ input "foo" }}`,
			},
			expectedError: "template: expression evaluator:1:4: executing \"expression evaluator\" at <input \"foo\">: error calling input: \"foo\" does not exist in the map of inputs",
		},
		{
			name: "lookup from previous outputs",
			previous: CommandOutputs{
				"step-1": map[string]any{
					"bar": "baz",
				},
			},
			local: With{
				"foo": `${{ from "step-1" "bar" }}`,
			},
			expectedTemplated: With{
				"foo": "baz",
			},
		},
		{
			name: "lookup from previous outputs - no outputs from step",
			local: With{
				"foo": `${{ from "step-1" "bar" }}`,
			},
			expectedError: `template: expression evaluator:1:4: executing "expression evaluator" at <from "step-1" "bar">: error calling from: no outputs for step "step-1"`,
		},
		{
			name: "lookup from previous outputs - missing arg",
			local: With{
				"foo": `${{ from "step-1" }}`,
			},
			expectedError: `template: expression evaluator:1:4: executing "expression evaluator" at <from>: wrong number of args for from: want 2 got 1`,
		},
		{
			name: "lookup from previous outputs - output from step not found",
			previous: CommandOutputs{
				"step-1": map[string]any{
					"bar": "baz",
				},
			},
			local: With{
				"foo": `${{ from "step-1" "dne" }}`,
			},
			expectedError: `template: expression evaluator:1:4: executing "expression evaluator" at <from "step-1" "dne">: error calling from: no output "dne" from "step-1"`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			templated, err := TemplateWith(t.Context(), tc.input, tc.local, tc.previous)
			if tc.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedError)
			}
			assert.Equal(t, tc.expectedTemplated, templated)
		})
	}
}
