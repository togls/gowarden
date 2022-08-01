package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/togls/gowarden/auth"
)

type IdentityHandler struct {
	logger *zerolog.Logger
	auth   auth.Authenticator
}

func NewIdentityHandler(
	logger *zerolog.Logger,
	auth auth.Authenticator,
) *IdentityHandler {
	return &IdentityHandler{
		logger: logger,
		auth:   auth,
	}
}

func (ih IdentityHandler) Routes(e *echo.Echo) {
	e.POST("/identity/connect/token", ih.login)
}

func (h IdentityHandler) login(c echo.Context) error {
	var cd auth.ConnectData
	if err := c.Bind(&cd); err != nil {
		return err
	}

	if err := cd.Validate(); err != nil {
		h.logger.Error().Fields(cd).Msg("validate connect data")
		return err
	}

	switch cd.GrantType {
	case auth.GTRefreshToken:
		data, err := h.auth.RefreshLogin(cd.RefreshToken)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, data)

	case auth.GTPassword:
		data, err := h.auth.PasswordLogin(&cd)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, data)

	default:
		// Invalid type
	}

	panic("todo")
}
