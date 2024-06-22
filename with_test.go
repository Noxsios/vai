// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPerformLookups(t *testing.T) {
	testCases := []struct {
		name              string
		input             With
		local             With
		previous          CommandOutputs
		expectedTemplated With
		expectedPersisted []string
		expectedError     string
	}{
		{
			name: "no lookups",
		},
		{
			name: "simple lookup + builtins",
			input: With{
				"key": "value",
			},
			local: With{
				"key":      "${{ input }}",
				"os":       "${{ .OS }}",
				"arch":     "${{ .ARCH }}",
				"platform": "${{ .PLATFORM }}",
			},
			expectedTemplated: With{
				"key":      "value",
				"os":       runtime.GOOS,
				"arch":     runtime.GOARCH,
				"platform": runtime.GOOS + "/" + runtime.GOARCH,
			},
		},
		{
			name: "lookup with defaults",
			input: With{
				"foo": "value",
			},
			local: With{
				"foo": "${{ input | default \"default\" }}",
				"bar": "${{ input | default \"default\" }}",
			},
			expectedTemplated: With{
				"foo": "value",
				"bar": "default",
			},
		},
		{
			name: "persisted lookups",
			input: With{
				"foo": "bar",
			},
			local: With{
				"foo": "${{ input | persist }}",
				"a":   "b${{ persist }}",
				"c":   "d",
				"e":   "${{ persist }}",
			},
			expectedPersisted: []string{"foo", "a", "e"},
			expectedTemplated: With{
				"foo": "bar",
				"a":   "b",
				"c":   "d",
				"e":   "",
			},
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
		{
			name: "invalid syntax",
			previous: CommandOutputs{
				"step-1": map[string]string{
					"bar": "baz",
				},
			},
			local: With{
				"foo": `${{ input`,
			},
			expectedError: `template: expression evaluator:1: unclosed action`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			templated, persisted, err := PerformLookups(tc.input, tc.local, tc.previous)
			if err != nil {
				require.EqualError(t, err, tc.expectedError)
			}
			require.Equal(t, tc.expectedTemplated, templated)
			require.ElementsMatch(t, tc.expectedPersisted, persisted)
		})
	}
}
