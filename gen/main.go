// Package main provides the entry point for the application.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/noxsios/vai"
	"github.com/noxsios/vai/cmd"
)

// main is the entry point for the application
func main() {
	logger := vai.Logger()
	schema := vai.WorkFlowSchema()

	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		logger.Fatal(err)
	}

	if err := os.WriteFile("vai.schema.json", b, 0644); err != nil {
		logger.Fatal(err)
	}

	if err := os.WriteFile("site/data/vai.schema.json", b, 0644); err != nil {
		logger.Fatal(err)
	}

	var buf bytes.Buffer
	rootCmd := cmd.NewRootCmd()
	rootCmd.SetOutput(&buf)
	if err = rootCmd.Usage(); err != nil {
		logger.Fatal(err)
	}

	usage := fmt.Sprintf("{\n  \"usage\": %q\n}", buf.String())

	if err := os.WriteFile("site/data/usage.json", []byte(usage), 0644); err != nil {
		logger.Fatal(err)
	}
}
