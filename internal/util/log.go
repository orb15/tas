package util

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const (
	LogLevelWarn     = "warn"
	LogLevelFlagName = "loglevel"
)

func NewLogger(loglevel ...string) *zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	logCfg := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	logger := zerolog.New(logCfg).With().Timestamp().Logger()

	if len(loglevel) > 0 {
		newLevel := strings.TrimSpace(strings.ToLower(loglevel[0]))
		switch newLevel {
		case "debug":
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		case "info":
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		case "warn":
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		case "error":
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		case "off":
			zerolog.SetGlobalLevel(zerolog.PanicLevel)
		default:
			fmt.Println()
			fmt.Printf("invalid logger level requested: %s, defaulting to ERROR\n", newLevel)
			fmt.Println()
		}
	}

	return &logger
}
