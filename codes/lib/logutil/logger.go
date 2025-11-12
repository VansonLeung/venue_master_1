package logutil

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// New builds a zerolog logger configured for the current environment.
func New(service string, env string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	writer := io.Writer(os.Stdout)
	if strings.ToLower(env) == "development" {
		writer = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen}
	}

	ctx := zerolog.New(writer).
		With().
		Timestamp().
		Str("service", service).
		Logger()

	level := strings.ToLower(os.Getenv("LOG_LEVEL"))
	switch level {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	return ctx
}
