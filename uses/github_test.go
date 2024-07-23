// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package uses

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGitHubFetcher(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests that require network access")
	}

	uses := "pkg:github/noxsios/vai@main?task=echo#testdata/simple.yaml"

	ctx := context.Background()

	client := NewGitHubClient()

	desc, err := client.Describe(ctx, uses)
	require.NoError(t, err)
	require.Equal(t, "53df01bd752c536a52836ccf988f656c3e4ed9d728aabed9974ac62453488840", desc.Hex)
	require.Equal(t, int64(122), desc.Size)

	rc, err := client.Fetch(ctx, uses)
	require.NoError(t, err)

	b, err := io.ReadAll(rc)
	require.NoError(t, err)

	require.Equal(t, `# yaml-language-server: $schema=../vai.schema.json

echo:
  - run: |
      echo "$MESSAGE"
    with:
      message: input
`, string(b))
}
