package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// New initialises the global zerolog logger and returns it.
// In development, output is human-readable (console). In production, JSON.
func New(appEnv string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	if appEnv == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.Kitchen})
	} else {
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	return log.Logger
}

// WithRequestID returns a child logger enriched with the given request ID.
func WithRequestID(logger zerolog.Logger, requestID string) zerolog.Logger {
	return logger.With().Str("request_id", requestID).Logger()
}

// WithUserID returns a child logger enriched with the given user ID.
func WithUserID(logger zerolog.Logger, userID string) zerolog.Logger {
	return logger.With().Str("user_id", userID).Logger()
}

// WithChapterID returns a child logger enriched with the given chapter ID.
func WithChapterID(logger zerolog.Logger, chapterID string) zerolog.Logger {
	return logger.With().Str("chapter_id", chapterID).Logger()
}
