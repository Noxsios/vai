package vai

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/invopop/jsonschema"
)

// Operation is an enum for the type of operation a step is performing
type Operation int

const (
	// OperationRun is a step that runs a command
	OperationRun Operation = iota
	// OperationUses is a step that calls another task
	OperationUses
)

// Step is a single step in a task
//
// While a step can have both `cmd` and `uses` fields, only one of them can be set
// at a time.
//
// This is enforced by JSON schema validation.
//
// TODO:
// - add `if` and `continue-on-error` fields?
// - add `timeout` field?
// - add `description` field?
type Step struct {
	// CMD is the command to run
	CMD    *string `json:"cmd,omitempty"`
	// Uses is a reference to a remote task
	Uses   *Uses   `json:"uses,omitempty"`
	// With is a map of additional parameters for the step/task call
	With   `json:"with,omitempty"`
	// Matrix is a matrix of parameters to run the step/task with
	Matrix `json:"matrix,omitempty"`
	// ID is a unique identifier for the step
	//
	// TODO: ensure this is unique in a given file and that it is a valid identifier
	ID string `json:"id,omitempty"`
}

// Operation returns the type of operation the step is performing
func (s Step) Operation() Operation {
	if s.CMD != nil {
		return OperationRun
	}
	if s.Uses != nil {
		return OperationUses
	}
	return -1
}

// Run executes the CMD field of a step
func (s Step) Run(with With, output *os.File) error {
	if s.CMD == nil {
		return fmt.Errorf("step does not have a command to run")
	}
	env := os.Environ()
	for k, v := range with {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	env = append(env, fmt.Sprintf("VAI_OUTPUT=%s", output.Name()))
	cmd := exec.Command("sh", "-e", "-c", *s.CMD)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	logger.Print(*s.CMD)
	return cmd.Run()
}

// JSONSchema returns the JSON schema for a step
//
// TODO: 
// - change this to _extend_ the schema, not provide the full schema
func (s Step) JSONSchema() *jsonschema.Schema {
	props := jsonschema.NewProperties()
	not := &jsonschema.Schema{
		Not: &jsonschema.Schema{},
	}
	props.Set("cmd", &jsonschema.Schema{
		Type:        "string",
		Description: "Command to run",
	})
	props.Set("uses", &jsonschema.Schema{
		Type:        "string",
		Description: "Call another task",
	})
	props.Set("id", &jsonschema.Schema{
		Type:        "string",
		Description: "Unique identifier for the step",
	})

	with := &jsonschema.Schema{
		Type:        "object",
		Description: "Additional parameters for the step/task call",
		AdditionalProperties: &jsonschema.Schema{
			OneOf: []*jsonschema.Schema{
				{
					Type: "string",
				},
				{
					Type: "boolean",
				},
				{
					Type: "integer",
				},
			},
		},
	}

	matrix := &jsonschema.Schema{
		Type:        "object",
		Description: "Matrix of parameters",
		AdditionalProperties: &jsonschema.Schema{
			Type: "array",
			Items: &jsonschema.Schema{
				OneOf: []*jsonschema.Schema{
					{
						Type: "string",
					},
					{
						Type: "boolean",
					},
					{
						Type: "integer",
					},
				},
			},
		},
	}

	props.Set("matrix", matrix)

	props.Set("with", with)

	cmdProps := jsonschema.NewProperties()
	cmdProps.Set("cmd", &jsonschema.Schema{
		Type: "string",
	})
	cmdProps.Set("uses", not)
	oneOfCmd := &jsonschema.Schema{
		Required:   []string{"cmd"},
		Properties: cmdProps,
	}

	usesProps := jsonschema.NewProperties()
	usesProps.Set("cmd", not)
	usesProps.Set("uses", &jsonschema.Schema{
		Type: "string",
	})
	oneOfUses := &jsonschema.Schema{
		Required:   []string{"uses"},
		Properties: usesProps,
	}

	return &jsonschema.Schema{
		Type:                 "object",
		Properties:           props,
		AdditionalProperties: jsonschema.FalseSchema,
		OneOf: []*jsonschema.Schema{
			oneOfCmd,
			oneOfUses,
		},
	}
}

// ParseOutputFile parses the output file of a step
func ParseOutputFile(f *os.File) (map[string]string, error) {
	_, err := f.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	// error if larger than 50MB, same limits as GitHub Actions
	if fi.Size() > 50*1024*1024 {
		return nil, fmt.Errorf("output file too large")
	}

	scanner := bufio.NewScanner(f)
	result := make(map[string]string)
	var currentKey string
	var currentDelimiter string
	var multiLineValue []string
	var collecting bool

	for scanner.Scan() {
		line := scanner.Text()

		if collecting {
			if line == currentDelimiter {
				// End of multiline value
				value := strings.Join(multiLineValue, "\n")
				result[currentKey] = value
				collecting = false
				multiLineValue = []string{}
				currentKey = ""
				currentDelimiter = ""
			} else {
				multiLineValue = append(multiLineValue, line)
			}
			continue
		}

		if idx := strings.Index(line, "="); idx != -1 {
			// Split the line at the first '=' to handle the key-value pair
			key := line[:idx]
			value := line[idx+1:]
			// Check if the value is a potential start of a multiline value
			if strings.HasSuffix(value, "<<") {
				currentKey = key
				currentDelimiter = strings.TrimSpace(value[2:])
				collecting = true
			} else {
				result[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Handle case where file ends but multiline was being collected
	if collecting && len(multiLineValue) > 0 {
		value := strings.Join(multiLineValue, "\n")
		result[currentKey] = value
	}

	return result, nil
}
