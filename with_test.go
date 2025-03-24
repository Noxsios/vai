// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package maru2

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
				"step-1": map[string]string{
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
				"step-1": map[string]string{
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
			if err != nil {
				require.EqualError(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expectedTemplated, templated)
		})
	}
}
