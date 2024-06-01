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

type LocalFetcher struct {
	fs afero.Fs
}

func NewLocalFetcher(fs afero.Fs) *LocalFetcher {
	return &LocalFetcher{fs}
}

func (f *LocalFetcher) Describe(_ context.Context, uses string) (Descriptor, error) {
	uri, err := url.Parse(uses)
	if err != nil {
		return Descriptor{}, err
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

func (f *LocalFetcher) Fetch(_ context.Context, uses string) (io.ReadCloser, error) {
	uri, err := url.Parse(uses)
	if err != nil {
		return nil, err
	}

	p := uri.Opaque
	return f.fs.Open(p)
}
