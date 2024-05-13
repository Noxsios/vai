package vai

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sync"

	"github.com/goccy/go-yaml"
	"github.com/package-url/packageurl-go"
	"github.com/xeipuuv/gojsonschema"
)

// TaskNamePattern is a regular expression for valid task names, it is also used for step IDs
var TaskNamePattern = regexp.MustCompile("^[_a-zA-Z][a-zA-Z0-9_-]*$")

// EnvVariablePattern is a regular expression for valid environment variable names
var EnvVariablePattern = regexp.MustCompile("^[a-zA-Z_]+[a-zA-Z0-9_]*$")

// Read reads a workflow from a file
func Read(filename string) (Workflow, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	wf := Workflow{}

	return wf, yaml.Unmarshal(b, &wf)
}

var schema string
var schemaOnce sync.Once

// Validate validates a workflow
func Validate(wf Workflow) error {
	for name, task := range wf {
		if ok := TaskNamePattern.MatchString(name); !ok {
			return fmt.Errorf("task name %q is invalid", name)
		}

		ids := make(map[string]int)

		for idx, step := range task {
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
				_, err := packageurl.FromString(UsesPrefix + step.Uses)
				if err != nil {
					return fmt.Errorf(".%s[%d].uses %q is invalid: %w", name, idx, step.Uses, err)
				}
			}

			if step.Uses != "" && step.Run != "" {
				return fmt.Errorf(".%s[%d] has both run and uses fields set", name, idx)
			}
		}
	}

	schemaOnce.Do(func() {
		s := WorkFlowSchema()
		b, err := json.Marshal(s)
		if err != nil {
			panic(err)
		}
		schema = string(b)
	})

	schemaLoader := gojsonschema.NewStringLoader(schema)

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
func ReadAndValidate(filename string) (Workflow, error) {
	wf, err := Read(filename)
	if err != nil {
		return nil, err
	}
	return wf, Validate(wf)
}
