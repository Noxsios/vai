// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package maru2

import (
	"bytes"
	"context"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/defenseunicorns/maru2/builtins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteBuiltin(t *testing.T) {
	testCases := []struct {
		name           string
		step           Step
		with           With
		dry            bool
		expectedError  string
		expectedLog    string
		expectedResult map[string]any
	}{
		{
			name: "echo builtin",
			step: Step{
				Uses: "builtin:echo",
				With: With{
					"text": "Hello, World!",
				},
			},
			with:           With{},
			dry:            false,
			expectedError:  "",
			expectedLog:    "Hello, World!\n",
			expectedResult: map[string]any{"stdout": "Hello, World!"},
		},
		{
			name: "echo builtin dry run",
			step: Step{
				Uses: "builtin:echo",
				With: With{
					"text": "Hello, World!",
				},
			},
			with:           With{},
			dry:            true,
			expectedError:  "",
			expectedLog:    "dry run",
			expectedResult: nil,
		},
		{
			name: "fetch builtin",
			step: Step{
				Uses: "builtin:fetch",
				With: With{
					"url":    "http://example.com",
					"method": "GET",
				},
			},
			with:           With{},
			dry:            true, // Use dry run to avoid actual HTTP requests
			expectedError:  "",
			expectedLog:    "dry run",
			expectedResult: nil,
		},
		{
			name: "non-existent builtin",
			step: Step{
				Uses: "builtin:nonexistent",
			},
			with:           With{},
			dry:            false,
			expectedError:  "builtin:nonexistent not found",
			expectedResult: nil,
		},
		{
			name: "echo builtin with invalid with",
			step: Step{
				Uses: "builtin:echo",
				With: With{
					"invalid": make(chan int), // Channels can't be marshaled to YAML
				},
			},
			with:           With{},
			dry:            false,
			expectedError:  "builtin:echo: [1:1] string was used where mapping is expected\n>  1 | <nil>\n       ^\n",
			expectedResult: nil,
		},
		{
			name: "fetch builtin with invalid with",
			step: Step{
				Uses: "builtin:fetch",
				With: With{
					"invalid": make(chan int), // Channels can't be marshaled to YAML
				},
			},
			with:           With{},
			dry:            false,
			expectedError:  "builtin:fetch: [1:1] string was used where mapping is expected\n>  1 | <nil>\n       ^\n",
			expectedResult: nil,
		},
		{
			name: "echo builtin with templated with",
			step: Step{
				Uses: "builtin:echo",
				With: With{
					"text": "${{ input \"greeting\" }}",
				},
			},
			with:           With{"greeting": "Hello from template"},
			dry:            false,
			expectedError:  "",
			expectedLog:    "Hello from template\n",
			expectedResult: map[string]any{"stdout": "Hello from template"},
		},
		{
			name: "echo builtin with broken structure",
			step: Step{
				Uses: "builtin:echo",
				With: With{
					"text": []string{"not", "a", "string"}, // Text should be a string, not an array
				},
			},
			with:           With{},
			dry:            false,
			expectedError:  "builtin:echo: json: cannot unmarshal array into Go struct field BuiltinEcho.text of type string",
			expectedResult: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := log.New(&buf)
			ctx := log.WithContext(context.Background(), logger)

			// Execute the builtin with the updated function signature
			result, err := ExecuteBuiltin(ctx, tc.step, tc.with, CommandOutputs{}, tc.dry)

			if tc.expectedError == "" {
				require.NoError(t, err)
				if tc.expectedResult != nil {
					assert.Equal(t, tc.expectedResult, result)
				}
			} else {
				require.EqualError(t, err, tc.expectedError)
				assert.Nil(t, result)
			}

			if tc.expectedLog != "" {
				assert.Contains(t, buf.String(), tc.expectedLog)
			}
		})
	}
}

func TestConvertWithToType(t *testing.T) {
	testCases := []struct {
		name          string
		with          With
		expectedValue any
		expectedError string
	}{
		{
			name: "convert to BuiltinEcho",
			with: With{
				"text": "Hello, World!",
			},
			expectedValue: builtins.BuiltinEcho{
				Text: "Hello, World!",
			},
			expectedError: "",
		},
		{
			name: "convert to BuiltinFetch",
			with: With{
				"url":     "http://example.com",
				"method":  "GET",
				"timeout": "30s",
				"headers": map[string]string{
					"Content-Type": "application/json",
				},
			},
			expectedValue: builtins.BuiltinFetch{
				URL:     "http://example.com",
				Method:  "GET",
				Timeout: "30s",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			expectedError: "",
		},
		{
			name: "convert with unmarshallable value",
			with: With{
				"invalid": make(chan int), // Channels can't be marshaled to JSON
			},
			expectedValue: builtins.BuiltinEcho{},
			expectedError: "json: unsupported type: chan int",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			switch expected := tc.expectedValue.(type) {
			case builtins.BuiltinEcho:
				result, err := ConvertWithTo[builtins.BuiltinEcho](tc.with)
				if tc.expectedError != "" {
					require.Error(t, err)
					assert.Contains(t, err.Error(), tc.expectedError)
				} else {
					require.NoError(t, err)
					assert.Equal(t, expected, result)
				}
			case builtins.BuiltinFetch:
				result, err := ConvertWithTo[builtins.BuiltinFetch](tc.with)
				if tc.expectedError != "" {
					require.Error(t, err)
					assert.Contains(t, err.Error(), tc.expectedError)
				} else {
					require.NoError(t, err)
					assert.Equal(t, expected, result)
				}
			}
		})
	}
}
