package vai

import (
	"os"

	"github.com/charmbracelet/log"
)

var logger = log.NewWithOptions(os.Stderr, log.Options{
	ReportTimestamp: false,
})

func Logger() *log.Logger {
	return logger
}

func SetLogLevel(level log.Level) {
	logger.SetLevel(level)
	if level == log.DebugLevel {
		logger.SetReportCaller(true)
	}
}
