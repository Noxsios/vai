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

type Fetcher[T any] func(context.Context, T) (io.ReadCloser, error)
type PackageURLFetcher Fetcher[packageurl.PackageURL]

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
	return resp.Body, nil
}

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
func ExecuteUses(ctx context.Context, store *Store, uses string, with With) error {
	logger.Debug("using", "task", uses)

	uri, err := url.Parse(uses)
	if err != nil {
		return err
	}

	if uri.Scheme == "" {
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

	switch uri.Scheme {
	case "http", "https":
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
		// TODO: handle paths from remote fetched workflows
		loc := uri.Opaque
		rc, err = FetchFile(loc)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown scheme: %s", uri.Scheme)
	}

	var wf Workflow

	if uri.Scheme == "file" {
		wf, err = ReadAndValidate(rc)
		if err != nil {
			return err
		}
	} else {
		b, err := io.ReadAll(rc)
		if err != nil {
			return err
		}
		exists, err := store.Exists(uses, bytes.NewReader(b))
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
			if err := store.Delete(uses); err != nil {
				return err
			}
		}

		if err := store.Store(uses, bytes.NewReader(b)); err != nil {
			return err
		}

		wf, err = store.Fetch(uses)
		if err != nil {
			return err
		}
	}

	var taskName string

	switch uri.Scheme {
	case "file":
		taskName = uri.Query().Get("task")
	default:
		taskName = pURL.Qualifiers.Map()["task"]
	}

	return Run(ctx, wf, taskName, with)
}
