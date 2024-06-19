// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"

	"github.com/noxsios/vai/storage"
	"github.com/package-url/packageurl-go"
)

// CacheEnvVar is the environment variable for the cache directory.
const CacheEnvVar = "VAI_CACHE"

// ExecuteUses runs a task from a remote workflow source.
func ExecuteUses(ctx context.Context, store *storage.Store, uses string, with With, prev string) error {
	logger.Debug("using", "task", uses)

	uri, err := url.Parse(uses)
	if err != nil {
		return err
	}

	if uri.Scheme == "" {
		return fmt.Errorf("must contain a scheme: %q", uses)
	}

	previous, err := url.Parse(prev)
	if err != nil {
		return err
	}

	if previous.Scheme == "" {
		return fmt.Errorf("must contain a scheme: %q", prev)
	}

	var next *url.URL

	if uri.Scheme == "file" {
		switch previous.Scheme {
		case "http", "https":
			// turn relative paths into absolute references
			next = previous
			next.Path = filepath.Join(filepath.Dir(previous.Path), uri.Opaque)
		case "pkg":
			pURL, err := packageurl.FromString(uses)
			if err != nil {
				return err
			}
			// turn relative paths into absolute references
			pURL.Subpath = filepath.Join(filepath.Dir(pURL.Subpath), uri.Opaque)
			next, _ = url.Parse(pURL.String())
		default:
			dir := filepath.Dir(previous.Opaque)
			if dir != "." {
				next = &url.URL{
					Scheme:   uri.Scheme,
					Opaque:   filepath.Join(dir, uri.Opaque),
					RawQuery: uri.RawQuery,
				}
				if next.Opaque == "." {
					next.Opaque = DefaultFileName
				}
			}
		}

		if next != nil {
			logger.Debug("merged", "previous", previous, "uses", uses, "next", next)
			uses = next.String()
		}
	}

	if next == nil {
		next, _ = url.Parse(uses)
	}

	if uri.Scheme == "pkg" {
		// dogsledding the error here since we know it's a package URL
		pURL, _ := packageurl.FromString(uses)
		if pURL.Subpath == "" {
			pURL.Subpath = DefaultFileName
		}
		if pURL.Version == "" {
			pURL.Version = "main"
		}
		uses = pURL.String()
	}

	fetcher, err := storage.SelectFetcher(uri, previous)
	if err != nil {
		return err
	}

	logger.Debug("chosen", "fetcher", fmt.Sprintf("%T", fetcher))

	var f io.ReadCloser

	if downloader, ok := fetcher.(storage.Downloader); ok {
		desc, err := downloader.Describe(ctx, uses)
		if err != nil {
			return err
		}

		exists, err := store.Exists(desc)
		if err != nil {
			return err
		}

		if !exists {
			logger.Debug("caching", "task", uses)
			rc, err := downloader.Fetch(ctx, uses)
			if err != nil {
				return err
			}
			defer rc.Close()

			if err := store.Store(rc); err != nil {
				return err
			}
		}

		f, err = store.Fetch(desc)
		if err != nil {
			return err
		}
		defer f.Close()
	} else {
		f, err = fetcher.Fetch(ctx, uses)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	if f == nil {
		return fmt.Errorf("failed to fetch %s referenced by %s", uses, prev)
	}

	wf, err := ReadAndValidate(f)
	if err != nil {
		return err
	}

	taskName := uri.Query().Get("task")

	return Run(ctx, store, wf, taskName, with, next.String())
}
