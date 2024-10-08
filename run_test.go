// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"context"
	"testing"
	"time"

	"github.com/noxsios/vai/uses"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	ctx := context.Background()
	fs := afero.NewMemMapFs()
	store, err := uses.NewStore(fs)
	require.NoError(t, err)
	with := With{}

	// simple happy path
	err = Run(ctx, store, helloWorldWorkflow, "", with, "file:test", false)
	require.NoError(t, err)

	// fast failure for 404
	err = Run(ctx, store, helloWorldWorkflow, "does not exist", with, "file:test", false)
	require.EqualError(t, err, "task \"does not exist\" not found")

	t.Run("fail on timeout - eval", func(t *testing.T) {
		ctx := context.TODO()
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		err := Run(ctx, store, Workflow{
			"timeout-eval": {Step{Eval: `
times := import("times")
times.sleep(3 * times.second)
`}},
		}, "timeout-eval", with, "file:test", false)
		require.EqualError(t, err, "context deadline exceeded")
	})

	t.Run("fail on timeout - run", func(t *testing.T) {
		ctx = context.TODO()
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		err = Run(ctx, store, Workflow{
			"timeout-run": {Step{Run: "sleep 3"}},
		}, "timeout-run", with, "file:test", false)
		require.EqualError(t, err, "signal: killed")
	})

	t.Run("boolean and int in with - eval", func(t *testing.T) {
		ctx = context.Background()
		err = Run(ctx, store, Workflow{
			"default": {Step{Eval: `
fmt := import("fmt")
fmt.printf("bool: %t, int: %d\n", b, i)
`, With: map[string]WithEntry{
				"b": true,
				"i": 42,
			}},
			},
		}, "", with, "file:test", false)
		require.NoError(t, err)
	})

	t.Run("boolean and int in with - run", func(t *testing.T) {
		ctx = context.Background()
		err = Run(ctx, store, Workflow{
			"default": {Step{Run: `
echo "bool: $B, int: $I"
`, With: map[string]WithEntry{
				"b": true,
				"i": 42,
			}},
			},
		}, "", with, "file:test", false)
		require.NoError(t, err)
	})
}

func TestToEnvVar(t *testing.T) {
	testCases := []struct {
		name     string
		s        string
		expected string
	}{
		{
			name: "empty",
		},
		{
			name:     "simple",
			s:        "foo",
			expected: "FOO",
		},
		{
			name:     "with dash",
			s:        "foo-bar",
			expected: "FOO_BAR",
		},
		{
			name:     "with underscore",
			s:        "foo_bar",
			expected: "FOO_BAR",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			actual := toEnvVar(tc.s)
			require.Equal(t, tc.expected, actual)
		})
	}
}
