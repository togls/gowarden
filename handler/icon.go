package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type IconHandler struct{}

func NewIconHandler() *IconHandler {
	return &IconHandler{}
}

func (ih *IconHandler) Routes(e *echo.Echo) {
	e.GET("/icons/:domain/icon.png", ih.GetIcon)
}

func (ih *IconHandler) GetIcon(c echo.Context) error {
	domain := c.Param("domain")

	if domain == "" {
		return c.NoContent(http.StatusBadRequest)
	}

	return c.Redirect(
		http.StatusMovedPermanently,
		"https://icons.bitwarden.net/"+domain+"/icon.png",
	)
}
