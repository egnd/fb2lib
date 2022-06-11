package factories

import (
	"io"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func NewZerolog(cfg *viper.Viper, writer io.Writer) zerolog.Logger {
	zerolog.DurationFieldUnit = cfg.GetDuration("logs.duration_unit")
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().In(time.Local)
	}

	if cfg.GetBool("logs.pretty") {
		writer = zerolog.ConsoleWriter{Out: writer,
			NoColor: !cfg.GetBool("logs.color"),
		}
	}

	logger := zerolog.New(writer).With().Timestamp().Logger().Level(zerolog.InfoLevel)

	if cfg.GetBool("logs.caller") {
		logger = logger.With().Caller().Logger()
	}

	if cfg.GetBool("logs.debug") {
		logger = logger.Level(zerolog.DebugLevel)
	}

	return logger
}
