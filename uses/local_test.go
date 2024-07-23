// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package uses

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestLocalFetcher(t *testing.T) {
	testCases := []struct {
		name string
		uses string
		desc Descriptor
		rc   io.ReadCloser

		expectedDescribeErr string
		expectedFetchErr    string
	}{
		{
			name: "file exists",
			uses: "file:foo.yaml",
			desc: Descriptor{Hex: "09ca7e4eaa6e8ae9c7d261167129184883644d07dfba7cbfbc4c8a2e08360d5b", Size: 12},
			rc:   io.NopCloser(strings.NewReader("hello, world")),
		},
		{
			name:                "file does not exist",
			uses:                "file:baz.yaml",
			expectedDescribeErr: "open baz.yaml: file does not exist",
			expectedFetchErr:    "open baz.yaml: file does not exist",
		},
		{
			name:                "invalid uri",
			uses:                "$%#",
			expectedDescribeErr: `parse "$%": invalid URL escape "%"`,
			expectedFetchErr:    `parse "$%": invalid URL escape "%"`,
		},
		{
			name:                "bad scheme",
			uses:                "http://foo.com/bar.yaml",
			expectedDescribeErr: `scheme is not "file"`,
			expectedFetchErr:    `scheme is not "file"`,
		},
	}

	fs := afero.NewMemMapFs()

	err := afero.WriteFile(fs, "foo.yaml", []byte("hello, world"), 0644)
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fetcher := NewLocalFetcher(fs)
			ctx := context.Background()

			desc, err := fetcher.Describe(ctx, tc.uses)
			if tc.expectedDescribeErr != "" {
				require.Equal(t, Descriptor{}, desc)
				require.EqualError(t, err, tc.expectedDescribeErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.desc, desc)
			}

			rc, err := fetcher.Fetch(ctx, tc.uses)
			if tc.expectedFetchErr != "" {
				require.Nil(t, rc)
				require.EqualError(t, err, tc.expectedFetchErr)
			} else {
				require.NoError(t, err)
				b1, err := io.ReadAll(tc.rc)
				require.NoError(t, err)

				b2, err := io.ReadAll(rc)
				require.NoError(t, err)

				require.Equal(t, string(b1), string(b2))
			}
		})
	}
}
