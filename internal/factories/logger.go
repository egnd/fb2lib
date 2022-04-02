package factories

import (
	"io"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func NewZerologLogger(cfg *viper.Viper, writer io.Writer) zerolog.Logger {
	zerolog.DurationFieldUnit = cfg.GetDuration("logs.duration_unit")
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().In(time.Local)
	}

	logger := zerolog.New(writer).With().Timestamp().
		Logger().Level(zerolog.InfoLevel)

	if cfg.GetBool("logs.debug") {
		logger = logger.Level(zerolog.DebugLevel)
	}

	if cfg.GetBool("logs.pretty") {
		logger = logger.Output(zerolog.ConsoleWriter{Out: writer})
	}

	return logger
}
