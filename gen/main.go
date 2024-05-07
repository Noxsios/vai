// Package main provides the entry point for the application.
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"text/template"

	"github.com/invopop/jsonschema"
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

	logger.Print("Generating vai.schema.json")

	if err := os.WriteFile("vai.schema.json", b, 0644); err != nil {
		logger.Fatal(err)
	}

	tmplContent, err := os.ReadFile("gen/docs.tmpl.md")
	if err != nil {
		logger.Fatal(err)
	}

	tmpl, err := template.New("docs").Delims("{%", "%}").Parse(string(tmplContent))
	if err != nil {
		logger.Fatal(err)
	}

	logger.Print("Generating docs/README.md")

	os.Remove("docs/README.md")

	f, err := os.Create("docs/README.md")
	if err != nil {
		logger.Fatal(err)
	}
	defer f.Close()

	var buf bytes.Buffer
	flags := cmd.NewRootCmd().Flags()
	flags.SetOutput(&buf)
	flags.PrintDefaults()

	meta := struct {
		Schema *jsonschema.Schema
		SchemaJSON string
		Flags string
	}{
		Schema: schema,
		Flags: buf.String(),
		SchemaJSON: string(b),
	}

	if err := tmpl.Execute(f, meta); err != nil {
		logger.Fatal(err)
	}
}
