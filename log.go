// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"context"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
)

// very side effect heavy
// should rethink this
func printScript(ctx context.Context, prefix, script string) {
	logger := log.FromContext(ctx)
	script = strings.TrimSpace(script)

	if termenv.EnvNoColor() {
		for _, line := range strings.Split(script, "\n") {
			logger.Printf("%s %s", prefix, line)
		}
		return
	}

	var buf strings.Builder
	style := "catppuccin-latte"
	if lipgloss.HasDarkBackground() {
		style = "catppuccin-frappe"
	}
	lang := "shell"
	if prefix == ">" {
		lang = "go"
	}
	if err := quick.Highlight(&buf, script, lang, "terminal256", style); err != nil {
		logger.Debugf("failed to highlight: %v", err)
		for _, line := range strings.Split(script, "\n") {
			logger.Printf("%s %s", prefix, line)
		}
		return
	}

	for _, line := range strings.Split(buf.String(), "\n") {
		logger.Printf("%s %s", prefix, line)
	}
}
