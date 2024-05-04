package vai

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/package-url/packageurl-go"
)

// UsesPrefix is the prefix for remote tasks.
const UsesPrefix = "pkg:vai/"

// FetchIntoStore fetches and stores a remote workflow into a given store.
func FetchIntoStore(pURL packageurl.PackageURL, store *Store) (Workflow, error) {
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

	resp, err := http.Get(raw)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	tmpFile, err := os.CreateTemp("", "vai-")
	if err != nil {
		return nil, err
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return nil, err
	}

	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	var wf Workflow

	b, err := io.ReadAll(tmpFile)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(b, &wf); err != nil {
		return nil, err
	}

	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	key := pURL.String()
	ok, err := store.Exists(key, tmpFile)
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

			_, err = tmpFile.Seek(0, 0)
			if err != nil {
				return nil, err
			}
			if err := store.Store(key, tmpFile); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else if !ok {
		_, err = tmpFile.Seek(0, 0)
		if err != nil {
			return nil, err
		}
		logger.Debug("caching", "workflow", pURL)
		if err = store.Store(key, tmpFile); err != nil {
			return nil, err
		}
	}

	return wf, nil
}

// RunUses runs a task from a remote workflow source.
func RunUses(uses string, with With) error {
	logger.Debug("using", "task", uses)

	pURL, err := packageurl.FromString(UsesPrefix + uses)
	if err != nil {
		return err
	}

	var wf Workflow

	if pURL.Subpath == "" {
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

		wf, err = ReadAndValidate(loc)
		if err != nil {
			return err
		}
	} else {
		store, err := DefaultStore()
		if err != nil {
			return err
		}
		wf, err = FetchIntoStore(pURL, store)
		if err != nil {
			return err
		}
	}

	taskName := pURL.Qualifiers.Map()["task"]

	return Run(wf, taskName, with)
}
