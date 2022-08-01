package handler

import (
	"net/http"

	"github.com/google/wire"
	"github.com/labstack/echo/v4"
	emd "github.com/labstack/echo/v4/middleware"

	"github.com/togls/gowarden/config"
	"github.com/togls/gowarden/handler/middleware"
)

var WireSet = wire.NewSet(
	NewMux,
	wire.Struct(new(MuxOptions), "*"),

	middleware.NewAppHeader,
	middleware.NewRecover,
	middleware.NewLogger,

	NewAccountHandler,
	NewCipherHandler,
	NewFolderHandler,
	NewOrganizationHandler,

	NewIdentityHandler,
	NewIconHandler,
)

type MuxOptions struct {
	Cfg *config.Core

	Account      *AccountHandler
	Cipher       *CipherHandler
	Folder       *FolderHandler
	Organization *OrganizationHandler
	Icon         *IconHandler
	Identity     *IdentityHandler

	AppHeader *middleware.AppHeader
	Recover   *middleware.Recover
	LoggerMW  *middleware.Logger
}

func NewMux(op *MuxOptions) http.Handler {

	e := echo.New()

	e.HTTPErrorHandler = op.Recover.HTTPErrorHandler

	e.Use(emd.CORS())
	e.Use(op.LoggerMW.Middleware)
	e.Use(op.AppHeader.Middleware)
	e.Use(op.Recover.Middleware)

	// Setup router
	rs := []Router{
		op.Account,
		op.Cipher,
		op.Folder,
		op.Organization,
		op.Icon,
		op.Identity,
	}
	for _, r := range rs {
		r.Routes(e)
	}

	e.Static("/", op.Cfg.Web)

	return e
}

type Router interface {
	Routes(e *echo.Echo)
}
