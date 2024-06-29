// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/muesli/termenv"
	"github.com/noxsios/vai/storage"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	ctx := context.Background()
	fs := afero.NewMemMapFs()
	store, err := storage.New(fs)
	require.NoError(t, err)
	with := With{}

	// simple happy path
	err = Run(ctx, store, helloWorldWorkflow, "", with, "file:test")
	require.NoError(t, err)

	// fast failure for 404
	err = Run(ctx, store, helloWorldWorkflow, "does not exist", with, "file:test")
	require.EqualError(t, err, "task \"does not exist\" not found")

	// fail on timeout - eval
	ctx = context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	err = Run(ctx, store, helloWorldWorkflow, "timeout-eval", with, "file:test")
	require.EqualError(t, err, "context deadline exceeded")

	// fail on timeout - run
	ctx = context.Background()
	ctx, cancel = context.WithTimeout(ctx, time.Second)
	defer cancel()
	err = Run(ctx, store, helloWorldWorkflow, "timeout-run", with, "file:test")
	require.EqualError(t, err, "context deadline exceeded")
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

func TestPrintScript(t *testing.T) {
	testCases := []struct {
		name     string
		script   string
		prefix   string
		expected string
	}{
		{
			name:     "simple eval",
			script:   `h := "hello"`,
			prefix:   ">",
			expected: "> h := \"hello\"\n",
		},
		{
			name:     "simple shell",
			script:   "echo hello",
			prefix:   "$",
			expected: "$ echo hello\n",
		},
		{
			name:     "multiline",
			script:   "echo hello\necho world\n\necho !",
			prefix:   "$",
			expected: "$ echo hello\n$ echo world\n$ echo !\n",
		},
	}

	var buf strings.Builder
	logger.SetOutput(&buf)
	_loggerColorProfile = termenv.Ascii
	t.Cleanup(func() {
		logger.SetOutput(os.Stderr)
		_loggerColorProfile = 0
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			printScript(tc.prefix, tc.script)
			require.Equal(t, tc.expected, buf.String())
			buf.Reset()
		})
	}
}
