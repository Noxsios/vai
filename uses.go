// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/package-url/packageurl-go"
)

// CacheEnvVar is the environment variable for the cache directory.
const CacheEnvVar = "VAI_CACHE"

// Force is a global flag to bypass SHA256 checksum verification for cached remote files.
var Force = false

// Fetcher is a generic fetcher function
type Fetcher[T any] func(context.Context, T) (io.ReadCloser, error)

// PackageURLFetcher is a fetcher for pURL
type PackageURLFetcher Fetcher[packageurl.PackageURL]

// GitHubFetcher is a PackageURLFetcher for GitHub
//
// If no subpath is given, the subpath is `vai.yaml`
//
// If no version is specified, the version is `main`
func GitHubFetcher(ctx context.Context, pURL packageurl.PackageURL) (io.ReadCloser, error) {
	if pURL.Subpath == "" {
		pURL.Subpath = DefaultFileName
	}

	if pURL.Version == "" {
		pURL.Version = "main"
	}

	raw := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", pURL.Namespace, pURL.Name, pURL.Version, pURL.Subpath)

	return FetchHTTP(ctx, raw)
}

// FetchHTTP performs a GET request using the default HTTP client
// against the provided raw URL string and returns the request body
//
// This function violates idiomatic Go's principle of not returning interfaces
// due to *http.Response.Body being directly typed as an io.ReadCloser
func FetchHTTP(ctx context.Context, raw string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "vai")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch %s: %s", raw, resp.Status)
	}
	return resp.Body, nil
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
func ExecuteUses(ctx context.Context, store *Store, uses string, with With, origin string) error {
	logger.Debug("using", "task", uses)

	usesURI, err := url.Parse(uses)
	if err != nil {
		return err
	}

	if usesURI.Scheme == "" {
		return fmt.Errorf("must contain a scheme: %s", uses)
	}

	var rc io.ReadCloser
	defer func() {
		if rc != nil {
			if err := rc.Close(); err != nil {
				logger.Warn(err)
			}
		}
	}()

	var pURL packageurl.PackageURL

	switch usesURI.Scheme {
	case "http", "https":
		// mutate the origin to the URL
		origin = uses
		rc, err = FetchHTTP(ctx, uses)
		if err != nil {
			return err
		}
	case "pkg":
		var fetch PackageURLFetcher
		pURL, err = packageurl.FromString(uses)
		if err != nil {
			return err
		}
		// mutate the origin to the package URL
		origin = uses
		switch pURL.Type {
		case "github":
			fetch = GitHubFetcher
		default:
			return fmt.Errorf("unsupported type: %s", pURL.Type)
		}

		rc, err = fetch(ctx, pURL)
		if err != nil {
			return err
		}
	case "file":
		loc := usesURI.Opaque

		originURL, err := url.Parse(origin)
		if err != nil {
			return err
		}

		switch originURL.Scheme {
		case "file":
			rc, err = FetchFile(loc)
			if err != nil {
				return err
			}
		case "http", "https":
			// turn relative paths into absolute references
			originURL.Path = filepath.Join(filepath.Dir(originURL.Path), loc)
			originURL.RawQuery = usesURI.RawQuery
			return ExecuteUses(ctx, store, originURL.String(), with, originURL.String())
		case "pkg":
			pURL, err = packageurl.FromString(uses)
			if err != nil {
				return err
			}
			// turn relative paths into absolute references
			pURL.Subpath = filepath.Join(filepath.Dir(pURL.Subpath), loc)
			return ExecuteUses(ctx, store, pURL.String(), with, pURL.String())
		}

	default:
		return fmt.Errorf("unknown scheme: %s", usesURI.Scheme)
	}

	var wf Workflow

	if usesURI.Scheme == "file" {
		wf, err = ReadAndValidate(rc)
		if err != nil {
			return err
		}
	} else {
		var key string
		// strip the task query parameter from the URL
		// to avoid caching the same workflow multiple times
		// TODO: make this a function
		switch usesURI.Scheme {
		case "pkg":
			// shallow copy to avoid modifying the original
			p := pURL
			p.Qualifiers = packageurl.Qualifiers{}
			key = p.String()
		default:
			shadowURI := *usesURI
			u := &shadowURI
			q := u.Query()
			q.Del("task")
			u.RawQuery = q.Encode()
			key = u.String()
		}

		fmt.Println("key", key)

		b, err := io.ReadAll(rc)
		if err != nil {
			return err
		}
		exists, err := store.Exists(key, bytes.NewReader(b))
		if err != nil && !IsHashMismatch(err) {
			return err
		}
		if err != nil && IsHashMismatch(err) && !Force {
			yes, err := ConfirmSHAOverwrite()
			if err != nil {
				return err
			}

			if !yes {
				return fmt.Errorf("hash mismatch, not overwriting")
			}
		}
		if exists {
			if err := store.Delete(key); err != nil {
				return err
			}
		}

		logger.Debug("caching", "task", key)
		if err := store.Store(key, bytes.NewReader(b)); err != nil {
			return err
		}

		wf, err = store.Fetch(key)
		if err != nil {
			return err
		}
	}

	var taskName string

	switch usesURI.Scheme {
	case "pkg":
		taskName = pURL.Qualifiers.Map()["task"]
	default:
		taskName = usesURI.Query().Get("task")
	}

	return Run(ctx, store, wf, taskName, with, origin)
}
