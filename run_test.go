// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"context"
	"testing"

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
