package vai

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

func Run(tg TaskGroup, outer With) error {
	use := func(u string, with With) error {
		logger.Debug("using", "task", u)

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

	for _, t := range tg {
		instances := make([]MatrixInstance, 0)
		for k, v := range t.Matrix {
			for _, i := range v {
				mi := make(MatrixInstance)
				mi[k] = i
				instances = append(instances, mi)
			}
		}
		if len(instances) == 0 {
			instances = append(instances, MatrixInstance{})
		}
		if t.Uses != nil && t.CMD != nil {
			return fmt.Errorf("task cannot have both cmd and uses")
		}
		for _, mi := range instances {
			w, err := PeformLookups(outer, t.With, mi)
			if err != nil {
				return err
			}

			if t.Uses != nil {
				if err := use(*t.Uses, w); err != nil {
					return err
				}
			}
			if t.CMD != nil {
				if err := t.Run(w); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
