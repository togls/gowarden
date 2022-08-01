package middleware

import (
	"net/http"
	"runtime"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/togls/gowarden/config"
)

type AppHeader struct {
	config.IframeAncestorsGetter
}

func NewAppHeader(getter config.IframeAncestorsGetter) *AppHeader {
	return &AppHeader{getter}
}

func (ah AppHeader) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Permissions-Policy", "accelerometer=(), ambient-light-sensor=(), autoplay=(), camera=(), encrypted-media=(), fullscreen=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), midi=(), payment=(), picture-in-picture=(), sync-xhr=(self \"https://haveibeenpwned.com\" \"https://2fa.directory\"), usb=(), vr=()")
		c.Response().Header().Set("Referrer-Policy", "same-origin")
		c.Response().Header().Set("X-Frame-Options", "SAMEORIGIN")
		c.Response().Header().Set("X-Content-Type-Options", "nosniff")
		c.Response().Header().Set("X-XSS-Protection", "1; mode=block")

		csp := "frame-ancestors 'self' chrome-extension://nngceckbapebfimnlniiiahkandclblb chrome-extension://jbkfoedolllekgbhcbcoahefnbanhhlh moz-extension://* " +
			ah.IframeAncestors() + ";"
		c.Response().Header().Set("Content-Security-Policy", csp)

		cc := c.Response().Header().Get("cache-control")
		if cc == "" {
			c.Response().Header().Set("cache-control", "no-cache, no-store, max-age=0")
		}
		return next(c)
	}
}

type Recover struct {
	logger *zerolog.Logger
}

func NewRecover(logger *zerolog.Logger) *Recover {
	return &Recover{logger}
}

func (rm *Recover) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func() {
			r := recover()
			if r == nil {
				return
			}

			err, ok := r.(error)
			if !ok {
				rm.logger.Error().Msgf("%v", r)
			}

			stack := make([]byte, 4<<10)
			length := runtime.Stack(stack, true)
			rm.logger.Error().Msgf("[PANIC RECOVER] %v %s\n", err, stack[:length])

			c.Error(err)
		}()
		return next(c)
	}
}

func (rm *Recover) HTTPErrorHandler(err error, c echo.Context) {

	if c.Response().Committed {
		return
	}

	he, ok := err.(*echo.HTTPError)
	if ok {
		if he.Internal != nil {
			if herr, ok := he.Internal.(*echo.HTTPError); ok {
				he = herr
			}
		}
	} else {
		he = &echo.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: http.StatusText(http.StatusInternalServerError),
		}
	}

	code := he.Code
	message := he.Message
	if m, ok := he.Message.(string); ok {
		message = HttpError{
			Message: m,
			Object:  "error",
			ErrorModel: ErrorModel{
				Message: m,
				Object:  "error",
			},
		}
	}

	// Send response
	if c.Request().Method == http.MethodHead { // Issue #608
		err = c.NoContent(he.Code)
	} else {
		err = c.JSON(code, message)
	}
	if err != nil {
		rm.logger.Debug().Err(err).Msg("failed to send response")
	}
}

type HttpError struct {
	Message               string     `json:"Message"`
	Error                 string     `json:"Error,omitempty"`
	Description           string     `json:"error_description"`
	ValidationErrors      []string   `json:"ValidationErrors,omitempty"`
	ErrorModel            ErrorModel `json:"ErrorModel,omitempty"`
	ExceptionMessage      any        `json:"ExceptionMessage"`
	ExceptionStackTrace   any        `json:"ExceptionStackTrace"`
	InnerExceptionMessage any        `json:"InnerExceptionMessage"`
	Object                string     `json:"Object,omitempty"`
}

type ErrorModel struct {
	Message string `json:"Message"`
	Object  string `json:"Object"`
}
