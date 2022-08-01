package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/togls/gowarden/config"

	_ "github.com/go-sql-driver/mysql"
)

func OpenDB(cfg *config.Core) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	return db, nil
}

type Apllication struct {
	server  *http.Server
	logger  *zerolog.Logger
	ctx     context.Context
	cleanup func()
}

type options struct {
	cfg     *config.Core
	handler http.Handler
	logger  *zerolog.Logger
	db      *sql.DB
}

func NewApplication(op options) *Apllication {

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)

	s := &http.Server{
		Addr:    op.cfg.Addr,
		Handler: op.handler,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	cleanup := func() {
		if err := op.db.Close(); err != nil {
			op.logger.Debug().Err(err).Msg("db close error")
		}

		cancel()
	}

	return &Apllication{
		server:  s,
		ctx:     ctx,
		cleanup: cleanup,
		logger:  op.logger,
	}
}

func (app *Apllication) Run() {
	defer app.cleanup()

	log := app.logger

	go func() {
		log.Info().Str("addr", app.server.Addr).Msg("http server started")

		if err := app.server.ListenAndServe(); err != nil {
			log.Err(err).Msg("server error")
		}
	}()

	<-app.ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.server.Shutdown(ctx); err != nil {
		log.Err(err).Msg("server shutdown error")
	}
}
