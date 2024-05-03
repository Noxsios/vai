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

// Uses is a reference to a remote task.
type Uses string

// String returns the string representation of a Uses.
func (u Uses) String() string {
	return string(u)
}

// Parse a Uses into a package URL.
func (u Uses) Parse() (packageurl.PackageURL, error) {
	return packageurl.FromString("pkg:vai/" + u.String())
}

// FetchIntoStore fetches and stores a remote workflow into a given store.
func FetchIntoStore(u Uses, store *Store) (Workflow, error) {
	// TODO: handle SHA provided within the URI so that we don't have to pull at all if we already have the file.
	pURL, err := u.Parse()
	if err != nil {
		return nil, err
	}

	// Example URL: "https://raw.githubusercontent.com/Noxsios/vai/main/tasks/echo.yaml"
	raw := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", pURL.Namespace, pURL.Name, pURL.Version, pURL.Subpath)

	if strings.Contains(raw, "//") {
		return nil, fmt.Errorf("invalid uses: %q", u)
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

	key := pURL.ToString()
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

// Run a Uses task.
func (u Uses) Run(with With) error {
	logger.Debug("using", "task", u)

	purl, err := u.Parse()
	if err != nil {
		return err
	}
	var wf Workflow

	if purl.Subpath == "" {
		var loc string

		switch purl.Namespace {
		case "", ".", "..":
			loc = purl.Name
		default:
			loc = filepath.Join(purl.Namespace, purl.Name)
		}

		fi, err := os.Stat(loc)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			loc = filepath.Join(loc, DefaultFileName)
		}

		b, err := os.ReadFile(loc)
		if err != nil {
			return err
		}

		if err := yaml.Unmarshal(b, &wf); err != nil {
			return err
		}
	} else {
		store, err := DefaultStore()
		if err != nil {
			return err
		}
		wf, err = FetchIntoStore(u, store)
		if err != nil {
			return err
		}
	}

	taskName := purl.Qualifiers.Map()["task"]

	return Run(wf, taskName, with)
}
