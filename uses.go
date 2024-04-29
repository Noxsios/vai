package vai

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"

	"github.com/goccy/go-yaml"
)

type Uses string

type URIComponents struct {
	Repository string
	FilePath   string
	TaskName   string
	Ref        string
}

func (u Uses) String() string {
	return string(u)
}

var URIRegex = regexp.MustCompile(`^(?:(?P<Repository>[^/.]+/[^/]+)(?:/))?(?P<FilePath>[^:@]+):(?P<TaskName>[^@]*)(?:@(?P<Ref>.+))?$`)

func (u Uses) Parse() (*URIComponents, error) {
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

func (u Uses) Fetch(store *Store) (Workflow, error) {
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

	ok, err := store.Exists(u.String(), tmpFile)
	if err != nil {
		if IsHashMismatch(err) {
			ok, err := ConfirmSHAOverwrite()
			if err != nil {
				return nil, err
			}

			if !ok {
				return nil, fmt.Errorf("hash mismatch, not overwriting")
			}

			if err := store.Delete(u.String()); err != nil {
				return nil, err
			}

			_, err = tmpFile.Seek(0, 0)
			if err != nil {
				return nil, err
			}
			if err := store.Store(u.String(), tmpFile); err != nil {
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
		logger.Debug("caching", "task", u)
		if err = store.Store(u.String(), tmpFile); err != nil {
			return nil, err
		}
	}

	return wf, nil
}

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

	tasks, err := wf.Find(uri.TaskName)
	if err != nil {
		return err
	}
	return Run(tasks, with)
}
