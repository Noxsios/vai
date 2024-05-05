// Package main provides the entry point for the application.
package main

import (
	"encoding/json"
	"os"

	"github.com/noxsios/vai"
)

// main is the entry point for the application
func main() {
	schema := vai.WorkFlowSchema()

	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		panic(err)
	}
	os.WriteFile("vai.schema.json", b, 0644)
}
