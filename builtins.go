// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package maru2

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/defenseunicorns/maru2/builtins"
)

// ExecuteBuiltin determines which builtin to run based upon the uses string, converts the With map to the expected struct, then calls the builtin's Execute method
func ExecuteBuiltin(ctx context.Context, uses string, with With, previous CommandOutputs, dry bool) (map[string]any, error) {
	name := strings.TrimPrefix(uses, "builtin:")
	logger := log.FromContext(ctx)

	builtinEmpty, ok := builtins.Builtins[name]
	if !ok {
		return nil, fmt.Errorf("%s not found", uses)
	}

	// what I'm doing here can't be legal
	var rendered With
	if with != nil {
		b, err := json.Marshal(with)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", uses, err)
		}
		templated, err := TemplateString(with, previous, string(b))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", uses, err)
		}
		if err := json.Unmarshal([]byte(templated), &rendered); err != nil {
			return nil, fmt.Errorf("%s: %w", uses, err)
		}
	}

	var builtin builtins.Builtin
	var err error
	switch builtinEmpty.(type) {
	case builtins.BuiltinEcho:
		builtin, err = ConvertWithTo[builtins.BuiltinEcho](rendered)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", uses, err)
		}
	case builtins.BuiltinFetch:
		builtin, err = ConvertWithTo[builtins.BuiltinFetch](rendered)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", uses, err)
		}
		// no default case due to map access handling that
	}

	if dry {
		logger.Info("dry run", "builtin", name, "with", rendered)
		return nil, nil
	}

	result, err := builtin.Execute(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", uses, err)
	}

	return result, nil
}

// ConvertWithTo transforms a With (map[string]any) to a Go struct through reparsing the map using generics
func ConvertWithTo[T any](with With) (T, error) {
	var result T

	b, err := json.Marshal(with)
	if err != nil {
		return result, err
	}

	return result, json.Unmarshal(b, &result)
}
