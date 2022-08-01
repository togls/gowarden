package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/togls/gowarden/config"
	"github.com/togls/gowarden/store"
)

type EmergencyAccessHandler struct {
	cfg *config.Core

	eas store.EmergencyAccess
}

func NewEmergencyAccessHandler(
	cfg *config.Core,
	eas store.EmergencyAccess,
) *EmergencyAccessHandler {
	return &EmergencyAccessHandler{
		cfg: cfg,

		eas: eas,
	}
}

func (eah EmergencyAccessHandler) Routes(e *echo.Echo) {
	// TODO: implement  EmergencyAccess
}
