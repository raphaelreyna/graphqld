package config

import (
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Log struct {
	Path  string
	JSON  bool
	Color bool
	Level zerolog.Level
}

func logFromMap(m map[string]interface{}) *Log {
	var logConf Log

	switch l := m["level"].(string); strings.ToLower(l) {
	case "info":
		logConf.Level = zerolog.InfoLevel
	case "warn":
		logConf.Level = zerolog.WarnLevel
	case "error":
		logConf.Level = zerolog.ErrorLevel
	case "fatal":
		logConf.Level = zerolog.FatalLevel
	case "disabled":
		logConf.Level = zerolog.Disabled
	default:
		if l != "" {
			log.Fatal().
				Str("value", l).
				Str("expected", "info | warn | error | fatal | disabled").
				Msg("invalid log level")
		}

		logConf.Level = zerolog.InfoLevel
	}

	if x, ok := m["path"].(string); ok {
		logConf.Path = x
	}

	if x, ok := m["json"].(bool); ok {
		logConf.JSON = x
	}

	if x, ok := m["color"].(bool); ok {
		logConf.Color = x
	}

	return &logConf
}
