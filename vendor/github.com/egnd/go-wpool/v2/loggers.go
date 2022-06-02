package wpool

import (
	"github.com/rs/zerolog"
)

// ZerologAdapter adapter for zerolog logger.
type ZerologAdapter struct {
	logger zerolog.Logger
}

// NewZerologAdapter creates adapter for zerolog logger.
func NewZerologAdapter(logger zerolog.Logger) *ZerologAdapter {
	return &ZerologAdapter{
		logger: logger,
	}
}

// Errorf logging error.
func (l ZerologAdapter) Errorf(err error, msg string, args ...interface{}) {
	l.logger.Error().Err(err).Msgf(msg, args...)
}

// Infof logging info.
func (l ZerologAdapter) Infof(msg string, args ...interface{}) {
	l.logger.Info().Msgf(msg, args...)
}
