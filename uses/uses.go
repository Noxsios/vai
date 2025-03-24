// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

// Package uses provides a cache+clients for storing and retrieving remote workflows.
package uses

import (
	"context"
	"io"
)

// Fetcher fetches a file from a remote location.
type Fetcher interface {
	Fetch(context.Context, string) (io.ReadCloser, error)
}

// Describer describes a file from a remote location.
type Describer interface {
	Describe(context.Context, string) (Descriptor, error)
}

// Downloader is a combination of a fetcher and a describer.
type Downloader interface {
	Fetcher
	Describer
}

// Descriptor describes a file to use for caching.
type Descriptor struct {
	Size int64
	Hex  string
}
