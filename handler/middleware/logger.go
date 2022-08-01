package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type Logger struct {
	logger *zerolog.Logger
}

func NewLogger(logger *zerolog.Logger) *Logger {
	l := logger.With().Logger()

	return &Logger{&l}
}

func (l Logger) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		logger := l.logger.Info()

		req := c.Request()
		res := c.Response()
		start := time.Now()
		if err = next(c); err != nil {
			logger = logger.Err(err)
			c.Error(err)
		}
		stop := time.Now()

		logger.Str("method", req.Method).
			Str("path", req.URL.Path).
			Int("status", res.Status).
			Dur("duration", stop.Sub(start)).
			Msg("")

		return
	}
}
