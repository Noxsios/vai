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
				"foo": "input ?? \"value\"",
				"bar": "input ?? \"default\"",
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
				"foo": "input | persist()",
				"c":   "'d'",
				"e":   `"" | persist()`,
			},
			expectedPersisted: []string{"foo", "e"},
			expectedTemplated: With{
				"foo": "bar",
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
				"foo": `from("step-1","bar")`,
			},
			expectedTemplated: With{
				"foo": "baz",
			},
		},
		{
			name: "lookup from previous outputs - no outputs from step",
			local: With{
				"foo": `from("step-1","bar")`,
			},
			expectedError: `no outputs for step "step-1" (1:1)
 | from("step-1","bar")
 | ^`,
		},
		{
			name: "lookup from previous outputs - output from step not found",
			previous: CommandOutputs{
				"step-1": map[string]string{
					"bar": "baz",
				},
			},
			local: With{
				"foo": `from("step-1","dne")`,
			},
			expectedError: `no output "dne" from "step-1" (1:1)
 | from("step-1","dne")
 | ^`,
		},
		{
			name: "no value to persist",
			local: With{
				"foo": `input | persist()`,
			},
			expectedError: `no value to persist (1:9)
 | input | persist()
 | ........^`,
		},
		{
			name: "invalid syntax",
			previous: CommandOutputs{
				"step-1": map[string]string{
					"bar": "baz",
				},
			},
			local: With{
				"foo": `input | persist`,
			},
			expectedError: `unexpected token EOF (1:15)
 | input | persist
 | ..............^`,
		},
		{
			name: "impossible - complex data structure",
			local: With{
				"foo": struct{ a string }{a: "bar"},
			},
			expectedError: `unsupported type struct { a string } for key "foo"`,
		},
		{
			name: "eval to nil",
			local: With{
				"foo": "input",
			},
			expectedError: `expression "input" evaluated to <nil>`,
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
