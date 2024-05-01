package vai

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/invopop/jsonschema"
)

type Operation int

const (
	OperationRun Operation = iota
	OperationUses
)

type Step struct {
	CMD    *string `json:"cmd,omitempty"`
	Uses   *Uses   `json:"uses,omitempty"`
	With   `json:"with,omitempty"`
	Matrix `json:"matrix,omitempty"`
	// TODO: ensure this is unique in a given file and that it is a valid identifier
	ID string `json:"id,omitempty"`
}

func (s Step) Operation() Operation {
	if s.CMD != nil {
		return OperationRun
	}
	if s.Uses != nil {
		return OperationUses
	}
	return -1
}

func (s Step) Run(with With, output *os.File) error {
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
