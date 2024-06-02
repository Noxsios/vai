// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package storage

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/url"

	"github.com/spf13/afero"
)

// LocalFetcher fetches a file from the local filesystem.
type LocalFetcher struct {
	fs afero.Fs
}

// NewLocalFetcher creates a new local fetcher
func NewLocalFetcher(fs afero.Fs) *LocalFetcher {
	return &LocalFetcher{fs}
}

// Describe returns a descriptor for the given file
func (f *LocalFetcher) Describe(_ context.Context, uses string) (Descriptor, error) {
	uri, err := url.Parse(uses)
	if err != nil {
		return Descriptor{}, err
	}

	if uri.Scheme != "file" {
		return Descriptor{}, fmt.Errorf("scheme is not \"file\"")
	}

	p := uri.Opaque

	fi, err := f.fs.Stat(p)
	if err != nil {
		return Descriptor{}, err
	}

	file, err := f.fs.Open(p)
	if err != nil {
		return Descriptor{}, err
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return Descriptor{}, err
	}

	return Descriptor{
		Size: fi.Size(),
		Hex:  fmt.Sprintf("%x", h.Sum(nil)),
	}, nil
}

// Fetch opens a file handle at the given location
func (f *LocalFetcher) Fetch(_ context.Context, uses string) (io.ReadCloser, error) {
	uri, err := url.Parse(uses)
	if err != nil {
		return nil, err
	}

	if uri.Scheme != "file" {
		return nil, fmt.Errorf("scheme is not \"file\"")
	}

	p := uri.Opaque
	return f.fs.Open(p)
}
