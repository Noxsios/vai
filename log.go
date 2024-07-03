// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"os"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
)

var logger = log.NewWithOptions(os.Stderr, log.Options{
	ReportTimestamp: false,
})

// Logger returns the global logger.
func Logger() *log.Logger {
	return logger
}

// SetLogLevel sets the global log level.
func SetLogLevel(level log.Level) {
	logger.SetLevel(level)
}

// very side effect heavy
// should rethink this
func printScript(prefix, script string) {
	script = strings.TrimSpace(script)

	if termenv.EnvNoColor() {
		for _, line := range strings.Split(script, "\n") {
			logger.Printf("%s %s", prefix, line)
		}
		return
	}

	customStyles := log.DefaultStyles()
	customStyles.Message = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#2f333a", Dark: "#d0d0d0"})
	logger.SetStyles(customStyles)
	defer logger.SetStyles(log.DefaultStyles())

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
