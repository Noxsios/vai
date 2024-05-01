package vai

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"
)

// Uses is a reference to a remote task.
type Uses string

// URIComponents contains the components of a URI.
type URIComponents struct {
	Repository string
	FilePath   string
	TaskName   string
	Ref        string
}

// String returns the string representation of a Uses.
func (u Uses) String() string {
	return string(u)
}

// URIRegex is a regular expression for parsing URIs.
//
// uses := Uses("Noxsios/vai/tasks/echo.yaml:world@main")
//
// components, _ := uses.Parse()
//
//	components == &URIComponents{
//		Repository: "Noxsios/vai",
//		FilePath:   "tasks/echo.yaml",
//		TaskName:   "world",
//		Ref:        "main",
//	}
var URIRegex = regexp.MustCompile(`^(?:(?P<Repository>[^/.]+/[^/]+)(?:/))?(?P<FilePath>[^:@]+):(?P<TaskName>[^@]*)(?:@(?P<Ref>.+))?$`)

// Parse parses a Uses URI.
func (u Uses) Parse() (*URIComponents, error) {
	// TODO: handle HTTPS URIs?

	matches := URIRegex.FindStringSubmatch(u.String())
	if matches == nil {
		return nil, fmt.Errorf("invalid uses URI: %s", u)
	}

	components := make(map[string]string)
	for i, name := range URIRegex.SubexpNames() {
		if i != 0 && name != "" {
			components[name] = matches[i]
		}
	}

	return &URIComponents{
		Repository: components["Repository"],
		FilePath:   components["FilePath"],
		TaskName:   components["TaskName"],
		Ref:        components["Ref"],
	}, nil
}

// Fetch fetches a remote workflow.
func (u Uses) Fetch(store *Store) (Workflow, error) {
	// TODO: handle SHA provided within the URI so that we don't have to pull at all if we already have the file.
	components, err := u.Parse()
	if err != nil {
		return nil, err
	}

	// Example URI: "Noxsios/vai/tasks/echo.yaml:world@main"
	// Example URL: "https://raw.githubusercontent.com/Noxsios/vai/main/tasks/echo.yaml"
	raw := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", components.Repository, components.Ref, components.FilePath)

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

	key := strings.Join([]string{components.Repository, components.Ref, components.FilePath}, "/")
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
		logger.Debug("caching", "workflow", u)
		if err = store.Store(key, tmpFile); err != nil {
			return nil, err
		}
	}

	return wf, nil
}

// Run runs a Uses task.
func (u Uses) Run(with With) error {
	logger.Debug("using", "task", u)

	uri, err := u.Parse()
	if err != nil {
		return err
	}
	logger.Debug("parsed", "uri", uri)
	var wf Workflow

	if uri.Repository == "" {
		b, err := os.ReadFile(uri.FilePath)
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
		wf, err = u.Fetch(store)
		if err != nil {
			return err
		}
	}

	return Run(wf, uri.TaskName, with)
}
