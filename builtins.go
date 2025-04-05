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
func ExecuteBuiltin(ctx context.Context, uses string, with With, dry bool) error {
	name := strings.TrimPrefix(uses, "builtin:")

	builtinEmpty, ok := builtins.Builtins[name]
	if !ok {
		return fmt.Errorf("builtin %q not found", name)
	}

	var builtin builtins.Builtin
	var err error
	switch builtinEmpty.(type) {
	case builtins.BuiltinEcho:
		builtin, err = ConvertWithToType[builtins.BuiltinEcho](with)
		if err != nil {
			return fmt.Errorf("builtin %q: %w", name, err)
		}
	case builtins.BuiltinFetch:
		builtin, err = ConvertWithToType[builtins.BuiltinFetch](with)
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

	err = builtin.Execute(ctx)
	if err != nil {
		return fmt.Errorf("builtin %q: %w", name, err)
	}

	return nil
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
