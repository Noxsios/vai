// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/go-github/v62/github"
	"github.com/package-url/packageurl-go"
)

// GitHubClient is a client for fetching files from GitHub
type GitHubClient struct {
	client *github.Client
}

// NewGitHubClient creates a new GitHub client
func NewGitHubClient() *GitHubClient {
	client := github.NewClient(nil)

	token, ok := os.LookupEnv("GITHUB_TOKEN")
	if ok {
		client = client.WithAuthToken(token)
	}
	return &GitHubClient{client}
}

// Describe returns a descriptor for the given file
func (g *GitHubClient) Describe(ctx context.Context, uses string) (Descriptor, error) {
	pURL, err := packageurl.FromString(uses)
	if err != nil {
		return Descriptor{}, err
	}

	fileContent, _, resp, err := g.client.Repositories.GetContents(ctx, pURL.Namespace, pURL.Name, pURL.Subpath, &github.RepositoryContentGetOptions{
		Ref: pURL.Version,
	})
	if err != nil {
		return Descriptor{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return Descriptor{}, fmt.Errorf("failed to get contents %s: %s", pURL, resp.Status)
	}

	if fileContent == nil {
		return Descriptor{}, fmt.Errorf("no content found for %s", pURL)
	}

	return Descriptor{
		Size: int64(fileContent.GetSize()),
		Hex:  fileContent.GetSHA(),
	}, nil
}

// Fetch fetches a file from a remote location
func (g *GitHubClient) Fetch(ctx context.Context, uses string) (io.ReadCloser, error) {
	pURL, err := packageurl.FromString(uses)
	if err != nil {
		return nil, err
	}

	rc, resp, err := g.client.Repositories.DownloadContents(ctx, pURL.Namespace, pURL.Name, pURL.Subpath, &github.RepositoryContentGetOptions{
		Ref: pURL.Version,
	})

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download %s: %s", pURL, resp.Status)
	}

	return rc, nil
}
