// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package storage

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

	uses := "pkg:github/noxsios/vai@main?task=hello-world#vai.yaml"

	ctx := context.Background()

	client := NewGitHubClient()

	desc, err := client.Describe(ctx, uses)
	require.NoError(t, err)
	require.Equal(t, desc.Hex, "ceb3c512fb9368eec89c66bef42378fd1e322c2f")
	require.Equal(t, desc.Size, int64(92))

	rc, err := client.Fetch(ctx, uses)
	require.NoError(t, err)

	b, err := io.ReadAll(rc)
	require.NoError(t, err)

	require.Equal(t, string(b), `# yaml-language-server: $schema=vai.schema.json

hello-world:
  - run: echo "Hello, World!"
`)
}
