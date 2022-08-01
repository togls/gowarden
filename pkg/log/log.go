package log

import (
	"time"

	"github.com/google/wire"
	"github.com/rs/zerolog"
)

var WireSet = wire.NewSet(
	New,
)

func New() *zerolog.Logger {

	out := zerolog.NewConsoleWriter()
	out.TimeFormat = time.RFC3339

	log := zerolog.New(out).
		Level(zerolog.DebugLevel).
		With().
		Timestamp().
		Logger()

	log.Info().Msg("logger initialized")

	return &log
}
