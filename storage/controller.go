// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package storage

import (
	"context"
	"io"
)

type Fetcher interface {
	Fetch(context.Context, string) (io.ReadCloser, error)
}

type Describer interface {
	Describe(context.Context, string) (Descriptor, error)
}

type Cacher interface {
	Fetcher
	Describer
}
