package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type LogFormat string

func (logFormat LogFormat) String() string {
	return string(logFormat)
}

const (
	JSONFormat    LogFormat = "json"
	ConsoleFormat LogFormat = "console"
)

// NewLogger returns a zerolog.Logger
func NewLogger(logLevel string, logFormat LogFormat) (zerolog.Logger, error) {
	level := zerolog.InfoLevel
	level, err := zerolog.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		return zerolog.Logger{}, err
	}
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = time.RFC3339Nano

	switch logFormat {
	case JSONFormat:
		return zerolog.New(os.Stdout).With().Caller().Timestamp().Logger(), nil
	case ConsoleFormat:
		return zerolog.New(os.Stdout).With().Caller().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stdout}), nil
	}

	return zerolog.Logger{}, fmt.Errorf("not support log format [%s]", logFormat)
}
