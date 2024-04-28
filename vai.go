package vai

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

func Run(tg TaskGroup, with With) error {
	use := func(u string, with With) error {
		fmt.Println(">", u)

		uri, err := url.Parse(u)
		if err != nil {
			return err
		}
		call := uri.Fragment

		if call == "" {
			return fmt.Errorf("no task call found")
		}

		if strings.Contains(call, "/") {
			return fmt.Errorf("task call cannot contain /")
		}

		if uri.Scheme != "" {
			// TODO: fetch
			return fmt.Errorf("scheme not supported")
		}

		var wf Workflow
		b, err := os.ReadFile(uri.Path)
		if err != nil {
			return err
		}

		if err := yaml.Unmarshal(b, &wf); err != nil {
			return err
		}

		tg, err := wf.Find(call)
		if err != nil {
			return err
		}
		return Run(tg, with)
	}

	executeAll := func(tg TaskGroup, with With) error {
		for _, t := range tg {
			if t.Uses != nil && t.CMD != nil {
				return fmt.Errorf("task cannot have both cmd and uses")
			}
			if t.Uses != nil {
				if err := t.With.Apply(with); err != nil {
					return err
				}
				return use(*t.Uses, t.With)
			}
			if t.CMD != nil {
				if t.With != nil {
					if err := t.With.Apply(with); err != nil {
						return err
					}
				}
				return t.Run()
			}
		}

		return nil
	}

	return executeAll(tg, with)
}
