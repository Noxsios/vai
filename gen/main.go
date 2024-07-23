// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

// Package main provides the entry point for the application.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/noxsios/vai"
	"github.com/noxsios/vai/cmd"
)

func run(root string) error {
	schema := vai.WorkFlowSchema()

	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(root, "vai.schema.json"), b, 0644); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(root, "site/data/vai.schema.json"), b, 0644); err != nil {
		return err
	}

	var buf bytes.Buffer
	rootCmd := cmd.NewRootCmd()
	rootCmd.SetOutput(&buf)
	if err = rootCmd.Usage(); err != nil {
		return err
	}

	usage := fmt.Sprintf("{\n  \"usage\": %q\n}", buf.String())

	return os.WriteFile(filepath.Join(root, "site/data/usage.json"), []byte(usage), 0644)
}

// main is the entry point for the application
func main() {
	// usage: `go run gen/main.go`
	if err := run(""); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
