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

func ExecuteBuiltin(ctx context.Context, uses string, with With, dry bool) error {
	name := strings.TrimPrefix(uses, "builtin:")

	builtin, ok := Builtins[name]
	if !ok {
		return fmt.Errorf("builtin %q not found", name)
	}

	if dry {
		logger := log.FromContext(ctx)
		logger.Info("dry run", "builtin", name)
		return nil
	}

	// TODO: how do we want to deal w/ conflicts between workflow inputs and builtin params
	withDefaults, err := MergeWithAndParams(ctx, with, builtin.Params)
	if err != nil {
		return err
	}

	return builtin.Execute(ctx, withDefaults)
}

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

type Builtin struct {
	Execute func(context.Context, With) error
	Params  InputMap
}

// Builtins maps builtin names to their implementations
var Builtins = map[string]Builtin{
	"echo": {
		Execute: Echo,
		Params: InputMap{
			"text": InputParameter{
				Description: "Text to echo",
				Required:    true,
			},
		},
	},
	"fetch": {
		Execute: Fetch,
		Params: InputMap{
			"url": InputParameter{
				Description: "URL to fetch",
				Required:    true,
			},
			"method": InputParameter{
				Description: "HTTP method to use",
				Required:    false,
				Default:     "GET",
			},
			"timeout": InputParameter{
				Description: "Timeout in seconds",
				Required:    false,
				Default:     30,
			},
		},
	},
}

// Echo is a builtin function that echoes text using log
func Echo(ctx context.Context, with With) error {
	logger := log.FromContext(ctx)
	text, ok := with["text"]
	if !ok {
		return fmt.Errorf("echo: missing required parameter 'text'")
	}

	logger.Print(text)
	return nil
}

// Fetch is a builtin function that makes HTTP requests
func Fetch(ctx context.Context, with With) error {
	logger := log.FromContext(ctx)
	urlParam, ok := with["url"]
	if !ok {
		return fmt.Errorf("fetch: missing required parameter 'url'")
	}

	url, ok := urlParam.(string)
	if !ok {
		return fmt.Errorf("fetch: 'url' parameter must be a string")
	}

	method := "GET"
	if methodParam, ok := with["method"]; ok {
		if methodStr, ok := methodParam.(string); ok {
			method = methodStr
		}
	}

	timeout := 30 * time.Second
	if timeoutParam, ok := with["timeout"]; ok {
		if timeoutInt, ok := timeoutParam.(int); ok {
			timeout = time.Duration(timeoutInt) * time.Second
		}
	}

	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
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
