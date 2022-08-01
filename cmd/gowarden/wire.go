//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/rs/zerolog"

	"github.com/togls/gowarden/auth"
	"github.com/togls/gowarden/config"
	"github.com/togls/gowarden/handler"
	"github.com/togls/gowarden/store/raw"
)

func createApp(configFile string, log *zerolog.Logger) (*Apllication, error) {
	panic(wire.Build(
		config.WireSet,
		handler.WireSet,
		auth.WireSet,
		raw.WireSet,

		OpenDB,
		NewApplication,
		wire.Struct(new(options), "*"),
	))
}
