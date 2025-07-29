package plugin

import (
	"github.com/hashicorp/go-hclog"
	"os"
)

func newLogger() hclog.Logger {
	return hclog.New(&hclog.LoggerOptions{
		Level:      newLogLevel(),
		Output:     os.Stderr,
		JSONFormat: true,
	})
}

func newLogLevel() hclog.Level {
	switch os.Getenv("JP_LOG") {
	case "TRACE":
		return hclog.Trace
	case "DEBUG":
		return hclog.Debug
	case "INFO":
		return hclog.Info
	case "WARN":
		return hclog.Warn
	case "ERROR":
		return hclog.Error
	default:
		return hclog.Off
	}
}
