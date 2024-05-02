package main

import (
	"encoding/json"
	"os"

	"github.com/Noxsios/vai"
	"github.com/invopop/jsonschema"
)

// main is the entry point for the application
func main() {
	reflector := jsonschema.Reflector{}
	reflector.ExpandedStruct = true
	schema := reflector.Reflect(&vai.Workflow{})

	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		panic(err)
	}
	os.WriteFile("vai.schema.json", b, 0644)
}
