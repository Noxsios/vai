package vai

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/package-url/packageurl-go"
)

const (
	// CacheEnvVar is the environment variable for the cache directory.
	CacheEnvVar = "VAI_CACHE"
	// UsesPrefix is the prefix for remote tasks.
	UsesPrefix = "pkg:vai/"
)

// Force is a global flag to bypass SHA256 checksum verification for cached remote files.
var Force = false

// FetchIntoStore fetches and stores a remote workflow into a given store.
func FetchIntoStore(ctx context.Context, pURL packageurl.PackageURL, store *Store) (Workflow, error) {
	// TODO: handle SHA provided within the URI so that we don't have to pull at all if we already have the file.

	if pURL.Subpath == "" {
		pURL.Subpath = DefaultFileName
	}

	var raw string

	switch ns := pURL.Namespace; {
	case strings.HasPrefix(ns, "github.com"):
		if pURL.Name == "" || pURL.Version == "" {
			return nil, fmt.Errorf("invalid uses: %#+v", pURL)
		}
		_, owner, ok := strings.Cut(ns, "github.com/")
		if !ok {
			return nil, fmt.Errorf("invalid uses: %q", pURL)
		}
		raw = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, pURL.Name, pURL.Version, pURL.Subpath)
	default:
		return nil, fmt.Errorf("unsupported namespace: %q", pURL.Namespace)
	}

	if strings.Contains(strings.TrimPrefix(raw, "https://"), "//") {
		return nil, fmt.Errorf("invalid uses: %#+v", pURL)
	}

	logger.Debug("fetching", "url", raw)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "vai")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	key := pURL.String()
	ok, err := store.Exists(key, bytes.NewReader(b))
	if err != nil {
		if IsHashMismatch(err) && !Force {
			ok, err := ConfirmSHAOverwrite()
			if err != nil {
				return nil, err
			}

			if !ok {
				return nil, fmt.Errorf("hash mismatch, not overwriting")
			}

			if err := store.Delete(key); err != nil {
				return nil, err
			}

			if err := store.Store(key, bytes.NewReader(b)); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else if !ok {
		logger.Debug("caching", "workflow", pURL)
		if err = store.Store(key, bytes.NewReader(b)); err != nil {
			return nil, err
		}
	}

	return store.Fetch(key)
}

// ExecuteUses runs a task from a remote workflow source.
func ExecuteUses(ctx context.Context, uses string, with With) error {
	logger.Debug("using", "task", uses)

	pURL, err := packageurl.FromString(UsesPrefix + uses)
	if err != nil {
		return err
	}

	var wf Workflow

	// If the pURL has no subpath or version, we assume it's a local file.
	if pURL.Subpath == "" && pURL.Version == "" {
		var loc string

		switch pURL.Namespace {
		case "", ".", "..":
			loc = pURL.Name
		default:
			loc = filepath.Join(pURL.Namespace, pURL.Name)
		}

		fi, err := os.Stat(loc)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			loc = filepath.Join(loc, DefaultFileName)
		}

		f, err := os.Open(loc)
		if err != nil {
			return err
		}
		defer f.Close()

		wf, err = ReadAndValidate(f)
		if err != nil {
			return err
		}
	} else {
		store, err := DefaultStore()
		if err != nil {
			return err
		}
		wf, err = FetchIntoStore(ctx, pURL, store)
		if err != nil {
			return err
		}
	}

	taskName := pURL.Qualifiers.Map()["task"]

	return Run(ctx, wf, taskName, with)
}
