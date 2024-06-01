// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/noxsios/vai/storage"
	"github.com/package-url/packageurl-go"
	"github.com/spf13/afero"
)

// CacheEnvVar is the environment variable for the cache directory.
const CacheEnvVar = "VAI_CACHE"

// DefaultStore creates a new store in the default location:
//
// $VAI_CACHE || $HOME/.vai/cache
func DefaultStore() (*storage.Store, error) {
	if cache := os.Getenv(CacheEnvVar); cache != "" {
		return storage.New(afero.NewBasePathFs(afero.NewOsFs(), cache))
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cache := filepath.Join(home, ".vai", "cache")

	return storage.New(afero.NewBasePathFs(afero.NewOsFs(), cache))
}

// FetchFile opens a file handle at the given location
//
// If the location is a directory, <loc>/vai.yaml is opened instead
//
// # This function is used to satisfy [io.ReadCloser] in other functions
//
// It is up to the caller to close the returned *os.File
func FetchFile(loc string) (*os.File, error) {
	fi, err := os.Stat(loc)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		loc = filepath.Join(loc, DefaultFileName)
	}
	f, err := os.Open(loc)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// ExecuteUses runs a task from a remote workflow source.
func ExecuteUses(ctx context.Context, store *storage.Store, uses string, with With, origin string) error {
	logger.Debug("using", "task", uses)

	uri, err := url.Parse(uses)
	if err != nil {
		return err
	}

	if uri.Scheme == "" {
		return fmt.Errorf("must contain a scheme: %s", uses)
	}

	var fetcher storage.Fetcher

	switch uri.Scheme {
	case "http", "https":
		// mutate the origin to the URL
		origin = uses
		fetcher = storage.NewHTTPFetcher()
	case "pkg":
		pURL, err := packageurl.FromString(uses)
		if err != nil {
			return err
		}
		if pURL.Subpath == "" {
			pURL.Subpath = DefaultFileName
		}
		if pURL.Version == "" {
			pURL.Version = "main"
		}

		uses = pURL.String()
		// mutate the origin to the URL
		origin = uses

		switch pURL.Type {
		case "github":
			fetcher = storage.NewGitHubClient()
		case "gitlab":
			fetcher, err = storage.NewGitLabClient(pURL.Qualifiers.Map()["base"])
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported type: %s", pURL.Type)
		}
	case "file":
		fetcher = storage.NewLocalFetcher(afero.NewOsFs())

		loc := uri.Opaque

		originURL, err := url.Parse(origin)
		if err != nil {
			return err
		}

		switch originURL.Scheme {
		case "http", "https":
			// turn relative paths into absolute references
			originURL.Path = filepath.Join(filepath.Dir(originURL.Path), loc)
			originURL.RawQuery = uri.RawQuery
			origin, uses = originURL.String(), originURL.String()
			fetcher = storage.NewHTTPFetcher()
		case "pkg":
			pURL, err := packageurl.FromString(uses)
			if err != nil {
				return err
			}
			// turn relative paths into absolute references
			pURL.Subpath = filepath.Join(filepath.Dir(pURL.Subpath), loc)
			origin, uses = pURL.String(), pURL.String()
			fetcher = storage.NewGitHubClient()
		default:
			dir := filepath.Dir(originURL.Opaque)
			if dir != "." {
				originURL.Opaque = filepath.Join(dir, loc)
				origin = originURL.String()
			}
		}

	default:
		return fmt.Errorf("unsupported scheme: %s", uri.Scheme)
	}

	logger.Debug("chosen", "fetcher", fmt.Sprintf("%T", fetcher))

	var f io.ReadCloser

	if cacher, ok := fetcher.(storage.Cacher); ok {
		desc, err := cacher.Describe(ctx, uses)
		if err != nil {
			return err
		}

		exists, err := store.Exists(desc)
		if err != nil {
			return err
		}

		if !exists {
			logger.Debug("caching", "task", uses)
			rc, err := cacher.Fetch(ctx, uses)
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
		return fmt.Errorf("failed to fetch %s referenced by %s", uses, origin)
	}

	wf, err := ReadAndValidate(f)
	if err != nil {
		return err
	}

	taskName := uri.Query().Get("task")

	return Run(ctx, store, wf, taskName, with, origin)
}
