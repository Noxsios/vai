// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/package-url/packageurl-go"
	"github.com/xanzy/go-gitlab"
)

type GitLabClient struct {
	client *gitlab.Client
}

func NewGitLabClient(base string) (*GitLabClient, error) {
	token := os.Getenv("GITLAB_TOKEN")
	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(base))
	if err != nil {
		return nil, err
	}

	return &GitLabClient{client}, nil
}

func (g *GitLabClient) Describe(ctx context.Context, uses string) (Descriptor, error) {
	pURL, err := packageurl.FromString(uses)
	if err != nil {
		return Descriptor{}, err
	}

	pid := pURL.Namespace + "/" + pURL.Name
	file, resp, err := g.client.RepositoryFiles.GetFileMetaData(pid, pURL.Subpath, &gitlab.GetFileMetaDataOptions{
		Ref: &pURL.Version,
	}, gitlab.WithContext(ctx))
	if err != nil {
		return Descriptor{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return Descriptor{}, fmt.Errorf("failed to get file metadata %s: %s", pURL, resp.Status)
	}

	return Descriptor{
		Size: int64(file.Size),
		Hex:  file.SHA256,
	}, nil
}

func (g *GitLabClient) Fetch(ctx context.Context, uses string) (io.ReadCloser, error) {
	pURL, err := packageurl.FromString(uses)
	if err != nil {
		return nil, err
	}

	pid := pURL.Namespace + "/" + pURL.Name
	b, resp, err := g.client.RepositoryFiles.GetRawFile(pid, pURL.Subpath, &gitlab.GetRawFileOptions{
		Ref: &pURL.Version,
	}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get file %s: %s", pURL, resp.Status)
	}

	return io.NopCloser(bytes.NewReader(b)), nil
}
