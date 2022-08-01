package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/togls/gowarden/store"
)

type DeviceHandler struct {
	devices store.Device
}

func (dh DeviceHandler) Routes(e *echo.Echo) {
	api := e.Group("/api")

	device := api.Group("/devices")
	device.PUT("/identifier/:uuid/clear-token", dh.ClearDeviceToken)
	device.PUT("/identifier/:uuid/token", dh.PutDeviceToken)
}

// clear_device_token
func (dh DeviceHandler) ClearDeviceToken(c echo.Context) error {
	id := c.Param("uuid")

	item, err := dh.devices.FindByUuid(id)
	if err != nil {
		return err
	}

	item.PushToken = nil

	if err := dh.devices.Save(item); err != nil {
		return err
	}

	return c.String(http.StatusOK, "")
}

// put_device_token
func (DeviceHandler) PutDeviceToken(c echo.Context) error {
	panic("TODO: not implemented")
}

// get_eq_domains
func (DeviceHandler) GetEqDomains(c echo.Context) error {
	panic("TODO: not implemented")
}

// post_eq_domains
func (DeviceHandler) PostEqDomains(c echo.Context) error {
	panic("TODO: not implemented")
}

// put_eq_domains
func (DeviceHandler) PutEqDomains(c echo.Context) error {
	panic("TODO: not implemented")
}

// hibp_breach
func (DeviceHandler) HibpBreach(c echo.Context) error {
	panic("TODO: not implemented")
}
