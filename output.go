package vai

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// CommandOutputs is a map of step IDs to their outputs.
//
// It is currently NOT goroutine safe.
type CommandOutputs map[string]map[string]string

// ParseOutput parses the output file of a step
func ParseOutput(r io.ReadSeeker) (map[string]string, error) {
	if f, ok := r.(*os.File); ok {
		fi, err := f.Stat()
		if err != nil {
			return nil, err
		}

		// error if larger than 50MB, same limits as GitHub Actions
		if fi.Size() > 50*1024*1024 {
			return nil, fmt.Errorf("output file too large")
		}
	}

	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(r)
	result := make(map[string]string)
	var currentKey, currentDelimiter string
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
