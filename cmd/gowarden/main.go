package main

import (
	"flag"

	"github.com/togls/gowarden/pkg/log"
)

func main() {
	var configFile string

	flag.StringVar(&configFile, "config", "", "config file")

	flag.Parse()

	logger := log.New()

	app, err := createApp(configFile, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create app")
	}

	app.Run()
}
