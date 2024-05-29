// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"os"

	"github.com/charmbracelet/log"
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
