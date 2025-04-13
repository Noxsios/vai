// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

package maru2

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/defenseunicorns/maru2/builtins"
)

// ExecuteBuiltin determines which builtin to run based upon the uses string, converts the With map to the expected struct, then calls the builtin's Execute method
func ExecuteBuiltin(ctx context.Context, step Step, with With, previous CommandOutputs, dry bool) (map[string]any, error) {
	name := strings.TrimPrefix(step.Uses, "builtin:")
	logger := log.FromContext(ctx)

	builtinEmpty, ok := builtins.Builtins[name]
	if !ok || builtinEmpty == nil {
		return nil, fmt.Errorf("%s not found", step.Uses)
	}

	var rendered With
	if with != nil {
		var err error
		rendered, err = TemplateWithMap(with, previous, step.With)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", step.Uses, err)
		}
	}

	builtinType := reflect.TypeOf(builtinEmpty)
	builtinValue := reflect.New(builtinType).Elem()

	if rendered != nil {
		data, err := json.Marshal(rendered)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", step.Uses, err)
		}

		if err := json.Unmarshal(data, builtinValue.Addr().Interface()); err != nil {
			return nil, fmt.Errorf("%s: %w", step.Uses, err)
		}
	}

	builtin, ok := builtinValue.Interface().(builtins.Builtin)
	if !ok {
		return nil, fmt.Errorf("%s: failed to convert to Builtin interface", step.Uses)
	}

	if dry {
		logger.Info("dry run", "builtin", name, "with", rendered)
		return nil, nil
	}

	result, err := builtin.Execute(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", step.Uses, err)
	}

	return result, nil
}
