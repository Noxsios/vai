// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package maru2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

// ExecuteBuiltin determines which builtin to run based upon the uses string, converts the With map to the expected struct, then calls the builtin's Execute method
func ExecuteBuiltin(ctx context.Context, uses string, with With, dry bool) error {
	name := strings.TrimPrefix(uses, "builtin:")

	builtinEmpty, ok := Builtins[name]
	if !ok {
		return fmt.Errorf("builtin %q not found", name)
	}

	var builtin Builtin
	var err error
	switch builtinEmpty.(type) {
	case BuiltinEcho:
		builtin, err = ConvertWithToType[BuiltinEcho](with)
		if err != nil {
			return fmt.Errorf("builtin %q: %w", name, err)
		}
	case BuiltinFetch:
		builtin, err = ConvertWithToType[BuiltinFetch](with)
		if err != nil {
			return fmt.Errorf("builtin %q: %w", name, err)
		}
	default:
		return fmt.Errorf("builtin %q not found", name)
	}

	if dry {
		logger := log.FromContext(ctx)
		logger.Info("dry run", "builtin", name)
		return nil
	}

	return builtin.Execute(ctx)
}

// MergeWithAndParams merges a With map into an InputMap, handling defaults, logging warnings on deprections, etc...
func MergeWithAndParams(ctx context.Context, with With, params InputMap) (With, error) {
	logger := log.FromContext(ctx)
	merged := maps.Clone(with)

	for name, param := range params {
		if _, ok := merged[name]; !ok {
			if param.Required && merged[name] == nil && param.Default == nil {
				return nil, fmt.Errorf("missing required input: %q", name)
			}
			if merged[name] == nil {
				merged[name] = param.Default
			}
			if param.DeprecatedMessage != "" && merged[name] != nil {
				logger.Warnf("input %q is deprecated: %s", name, param.DeprecatedMessage)
			}
		}
	}

	return merged, nil
}

// Builtin is a simple interface, only implementable on structs due to how the with re-parsing logic works
type Builtin interface {
	Execute(ctx context.Context) error
}

// ConvertWithToType transforms a With (map[string]any) to a Go struct through reparsing the map using generics
func ConvertWithToType[T any](with With) (T, error) {
	var result T

	b, err := json.Marshal(with)
	if err != nil {
		return result, err
	}

	return result, json.Unmarshal(b, &result)
}

// Builtins maps builtin names to their implementations
var Builtins = map[string]Builtin{
	"echo":  BuiltinEcho{},
	"fetch": BuiltinFetch{},
}

// BuiltinEcho is a sample builtin to MVP execution
type BuiltinEcho struct {
	Text string `json:"text" jsonschema:"description=Text to echo"`
}

// Execute the builtin
func (b BuiltinEcho) Execute(ctx context.Context) error {
	logger := log.FromContext(ctx)

	logger.Print(b.Text)
	return nil
}

// BuiltinFetch is a sample builtin to showcase configuration and schema gen
type BuiltinFetch struct {
	URL    string `json:"url" jsonschema:"description=URL to fetch"`
	Method string `json:"method,omitempty" jsonschema:"description=HTTP method to use"`
	// TODO: this is time in nanoseconds
	Timeout time.Duration `json:"timeout,omitempty" jsonschema:"description=Timeout for the request"`

	Headers map[string]string `json:"headers,omitempty" jsonschema:"description=HTTP headers to send"`
}

// Execute the builtin
func (b BuiltinFetch) Execute(ctx context.Context) error {
	logger := log.FromContext(ctx)

	method := b.Method
	if method == "" {
		method = "GET"
	}

	// timeout := b.Timeout
	timeout := 30 * time.Second

	client := &http.Client{
		Timeout: timeout,
	}

	logger.Printf("Headers: %s", b.Headers)

	req, err := http.NewRequestWithContext(ctx, method, b.URL, nil)
	if err != nil {
		return fmt.Errorf("fetch: error creating request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch: error executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("fetch: error reading response body: %w", err)
	}

	logger.Printf("Status: %s", resp.Status)
	logger.Printf("Content-Type: %s", resp.Header.Get("Content-Type"))
	logger.Printf("Content-Length: %d", len(body))

	if resp.Header.Get("Content-Type") == "application/json" {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
			logger.Print("Response Body:")
			logger.Print(prettyJSON.String())
			return nil
		}
	}

	logger.Print("Response Body:")
	logger.Print(string(body))

	return nil
}
