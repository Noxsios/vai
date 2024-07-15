// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"context"
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
				"key":      "input",
				"os":       "os",
				"arch":     "arch",
				"platform": "platform",
				"boolean":  true,
				"int":      1,
			},
			expectedTemplated: With{
				"key":      "value",
				"os":       runtime.GOOS,
				"arch":     runtime.GOARCH,
				"platform": runtime.GOOS + "/" + runtime.GOARCH,
				"boolean":  true,
				"int":      1,
			},
		},
		{
			name: "lookup with defaults",
			input: With{
				"foo": "value",
			},
			local: With{
				"foo": "input || \"value\"",
				"bar": "input || \"default\"",
			},
			expectedTemplated: With{
				"foo": "value",
				"bar": "default",
			},
		},
		{
			name: "lookup from previous outputs",
			previous: CommandOutputs{
				"step-1": map[string]any{
					"bar": "baz",
				},
			},
			local: With{
				"foo": `steps["step-1"].bar`,
			},
			expectedTemplated: With{
				"foo": "baz",
			},
		},
		{
			name: "lookup from previous outputs - no outputs from step",
			local: With{
				"foo": `steps["step-1"].bar`,
			},
			expectedError: "expression evaluated to <nil>:\n\tsteps[\"step-1\"].bar",
		},
		{
			name: "lookup from previous outputs - output from step not found",
			previous: CommandOutputs{
				"step-1": map[string]any{
					"bar": "baz",
				},
			},
			local: With{
				"foo": `steps["step-1"].dne`,
			},
			expectedError: "expression evaluated to <nil>:\n\tsteps[\"step-1\"].dne",
		},
		{
			name: "invalid syntax",
			previous: CommandOutputs{
				"step-1": map[string]any{
					"bar": "baz",
				},
			},
			local: With{
				"foo": `input | persist`,
			},
			expectedError: "script run: Compile Error: unresolved reference 'persist'\n\tat (main):1:21",
		},
		{
			name: "eval to nil",
			local: With{
				"foo": "input",
			},
			expectedError: "expression evaluated to <nil>:\n\tinput",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			templated, err := PerformLookups(context.TODO(), tc.input, tc.local, tc.previous)
			if err != nil {
				require.EqualError(t, err, tc.expectedError)
			}
			require.Equal(t, tc.expectedTemplated, templated)
		})
	}
}
