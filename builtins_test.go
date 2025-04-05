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
		name          string
		uses          string
		with          With
		dry           bool
		expectedError string
		expectedLog   string
	}{
		{
			name: "echo builtin",
			uses: "builtin:echo",
			with: With{
				"text": "Hello, World!",
			},
			dry:           false,
			expectedError: "",
			expectedLog:   "Hello, World!\n",
		},
		{
			name: "echo builtin dry run",
			uses: "builtin:echo",
			with: With{
				"text": "Hello, World!",
			},
			dry:           true,
			expectedError: "",
			expectedLog:   "dry run",
		},
		{
			name: "fetch builtin",
			uses: "builtin:fetch",
			with: With{
				"url":    "http://example.com",
				"method": "GET",
			},
			dry:           true, // Use dry run to avoid actual HTTP requests
			expectedError: "",
			expectedLog:   "dry run",
		},
		{
			name:          "non-existent builtin",
			uses:          "builtin:nonexistent",
			with:          With{},
			dry:           false,
			expectedError: "builtin \"nonexistent\" not found",
		},
		{
			name: "echo builtin with invalid with",
			uses: "builtin:echo",
			with: With{
				"invalid": make(chan int), // Channels can't be marshaled to JSON
			},
			dry:           false,
			expectedError: "builtin \"echo\": json: unsupported type: chan int",
		},
		{
			name: "fetch builtin with invalid with",
			uses: "builtin:fetch",
			with: With{
				"invalid": make(chan int), // Channels can't be marshaled to JSON
			},
			dry:           false,
			expectedError: "builtin \"fetch\": json: unsupported type: chan int",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := log.New(&buf)
			ctx := log.WithContext(context.Background(), logger)

			err := ExecuteBuiltin(ctx, tc.uses, tc.with, tc.dry)

			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
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
		expectedValue interface{}
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
				result, err := ConvertWithToType[builtins.BuiltinEcho](tc.with)
				if tc.expectedError != "" {
					require.Error(t, err)
					assert.Contains(t, err.Error(), tc.expectedError)
				} else {
					require.NoError(t, err)
					assert.Equal(t, expected, result)
				}
			case builtins.BuiltinFetch:
				result, err := ConvertWithToType[builtins.BuiltinFetch](tc.with)
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
