// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package builtins

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
)

// Builtin is a simple interface, only implementable on structs due to how the with re-parsing logic works
type Builtin interface {
	Execute(ctx context.Context) error
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
	URL     string        `json:"url" jsonschema:"description=URL to fetch"`
	Method  string        `json:"method,omitempty" jsonschema:"description=HTTP method to use"`
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

	timeout := b.Timeout * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

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
