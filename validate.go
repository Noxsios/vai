// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/goccy/go-yaml"
	"github.com/xeipuuv/gojsonschema"
)

// TaskNamePattern is a regular expression for valid task names, it is also used for step IDs
var TaskNamePattern = regexp.MustCompile("^[_a-zA-Z][a-zA-Z0-9_-]*$")

// EnvVariablePattern is a regular expression for valid environment variable names
var EnvVariablePattern = regexp.MustCompile("^[a-zA-Z_]+[a-zA-Z0-9_]*$")

// Read reads a workflow from a file
func Read(r io.Reader) (Workflow, error) {
	if rs, ok := r.(io.Seeker); ok {
		_, err := rs.Seek(0, io.SeekStart)
		if err != nil {
			return nil, err
		}
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	wf := Workflow{}

	return wf, yaml.Unmarshal(b, &wf)
}

var _schema string
var _schemaOnce sync.Once

// Validate validates a workflow
func Validate(wf Workflow) error {
	for name, task := range wf {
		if ok := TaskNamePattern.MatchString(name); !ok {
			return fmt.Errorf("task name %q is invalid", name)
		}

		ids := make(map[string]int, len(task))

		for idx, step := range task {
			// ensure that only one of run or uses or eval fields is set
			// if more than one is set, return an error
			// if none are set, return an error
			switch {
			case step.Uses != "" && step.Run != "":
				return fmt.Errorf(".%s[%d] has both run and uses fields set", name, idx)
			case step.Uses != "" && step.Eval != "":
				return fmt.Errorf(".%s[%d] has both eval and uses fields set", name, idx)
			case step.Run != "" && step.Eval != "":
				return fmt.Errorf(".%s[%d] has both run and eval fields set", name, idx)
			case step.Uses == "" && step.Run == "" && step.Eval == "":
				return fmt.Errorf(".%s[%d] must have one of [eval, run, uses] fields set", name, idx)
			}

			if step.ID != "" {
				if ok := TaskNamePattern.MatchString(step.ID); !ok {
					return fmt.Errorf(".%s[%d].id %q is invalid", name, idx, step.ID)
				}

				if _, ok := ids[step.ID]; ok {
					return fmt.Errorf(".%s[%d] and .%s[%d] have the same ID %q", name, ids[step.ID], name, idx, step.ID)
				}
				ids[step.ID] = idx
			}

			if step.Uses != "" {
				u, err := url.Parse(step.Uses)
				if err != nil {
					return fmt.Errorf(".%s[%d].uses %q is invalid", name, idx, step.Uses)
				}

				if u.Scheme == "" {
					// if step.Uses == name {
					// 	return fmt.Errorf(".%s[%d].uses cannot reference itself", name, idx)
					// }
					_, ok := wf.Find(step.Uses)
					if !ok {
						return fmt.Errorf(".%s[%d].uses %q not found", name, idx, step.Uses)
					}
				} else {
					schemes := []string{"file", "http", "https", "pkg"}

					if !slices.Contains(schemes, u.Scheme) {
						return fmt.Errorf(".%s[%d].uses %q is not one of [%s]", name, idx, u.Scheme, strings.Join(schemes, ", "))
					}
				}
			}
		}
	}

	_schemaOnce.Do(func() {
		s := WorkFlowSchema()
		b, err := json.Marshal(s)
		if err != nil {
			panic(err)
		}
		_schema = string(b)
	})

	schemaLoader := gojsonschema.NewStringLoader(_schema)

	result, err := gojsonschema.Validate(schemaLoader, gojsonschema.NewGoLoader(wf))
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	}

	for _, err := range result.Errors() {
		logger.Error(err.String())
	}

	return fmt.Errorf("schema validation failed")
}

// ReadAndValidate reads and validates a workflow
func ReadAndValidate(r io.Reader) (Workflow, error) {
	wf, err := Read(r)
	if err != nil {
		return nil, err
	}
	return wf, Validate(wf)
}
