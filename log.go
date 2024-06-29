// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package vai

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
)

var logger = log.NewWithOptions(os.Stderr, log.Options{
	ReportTimestamp: false,
})

var _loggerColorProfile termenv.Profile

// Logger returns the global logger.
func Logger() *log.Logger {
	return logger
}

// SetLogLevel sets the global log level.
func SetLogLevel(level log.Level) {
	logger.SetLevel(level)
}

// SetColorProfile sets the global color profile.
func SetColorProfile(p termenv.Profile) {
	_loggerColorProfile = p
	logger.SetColorProfile(p)
}
