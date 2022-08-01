package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/togls/gowarden/auth"
	"github.com/togls/gowarden/config"
	"github.com/togls/gowarden/model"
	"github.com/togls/gowarden/store"
)

type FolderHandler struct {
	folders store.Folder
	auth    *auth.Core
	logger  *zerolog.Logger
}

func NewFolderHandler(
	cfg *config.Core,
	folders store.Folder,
	auth *auth.Core,
) *FolderHandler {
	return &FolderHandler{
		folders: folders,
		auth:    auth,
		logger:  cfg.Logger,
	}
}

func (fh *FolderHandler) Routes(e *echo.Echo) {
	folder := e.Group("/api/folders", fh.auth.RequireAuth)

	folder.GET("", fh.GetFolders)
	folder.GET("/:uuid", fh.GetFolder)
	folder.POST("", fh.PostFolders)
	folder.POST("/:uuid", fh.PostFolder)
	folder.PUT("/:uuid", fh.PutFolder)
	folder.POST("/:uuid/delete", fh.DeleteFolderPost)
	folder.DELETE("/:uuid", fh.DeleteFolder)
}

func (fh *FolderHandler) GetFolders(c echo.Context) error {
	user := auth.GetUser(c)

	folders, err := fh.folders.FindByUser(user.Uuid)
	if err != nil {
		return err
	}

	resp := struct {
		Data              []*model.Folder `json:"Data"`
		Object            string          `json:"Object"`
		ContinuationToken any             `json:"ContinuationToken"`
	}{
		Data:              folders,
		Object:            "list",
		ContinuationToken: nil,
	}

	return c.JSON(200, resp)
}

func (fh *FolderHandler) GetFolder(c echo.Context) error {
	fUuid := c.Param("uuid")

	user := auth.GetUser(c)

	folder, err := fh.folders.FindByUuid(fUuid)
	if err != nil {
		return err
	}

	if folder.UserUuid != user.Uuid {
		return echo.NewHTTPError(http.StatusBadRequest, "Folder belongs to another user")
	}

	return c.JSON(http.StatusOK, folder)
}

type FolderData struct {
	Name string `json:"Name"`
}

func (fd FolderData) toFolder(user string) (*model.Folder, error) {
	ur, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &model.Folder{
		Uuid:     ur.String(),
		Name:     fd.Name,
		UserUuid: user,
	}, nil
}

func (fh *FolderHandler) PostFolders(c echo.Context) error {
	data := new(FolderData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)

	folder, err := data.toFolder(user.Uuid)
	if err != nil {
		return err
	}

	if err := fh.folders.Create(folder); err != nil {
		fh.logger.Debug().Err(err).
			Str("user", user.Uuid).
			Str("name", folder.Name).
			Msg("")
		return err
	}

	return c.JSON(http.StatusOK, folder)
}

func (fh *FolderHandler) PostFolder(c echo.Context) error {
	return fh.PutFolder(c)
}

func (fh *FolderHandler) PutFolder(c echo.Context) error {
	data := new(FolderData)

	if err := c.Bind(data); err != nil {
		return err
	}

	fUuid := c.Param("uuid")

	user := auth.GetUser(c)

	folder, err := fh.folders.FindByUuid(fUuid)
	if err != nil {
		return err
	}

	if folder.UserUuid != user.Uuid {
		return echo.NewHTTPError(http.StatusBadRequest, "Folder belongs to another user")
	}

	folder.Name = data.Name

	if err := fh.folders.Create(folder); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, folder)
}

func (fh *FolderHandler) DeleteFolderPost(c echo.Context) error {
	return fh.DeleteFolder(c)
}

func (fh *FolderHandler) DeleteFolder(c echo.Context) error {
	fUuid := c.Param("uuid")

	user := auth.GetUser(c)

	folder, err := fh.folders.FindByUuid(fUuid)
	if err != nil {
		return err
	}

	if folder.UserUuid != user.Uuid {
		return echo.NewHTTPError(http.StatusBadRequest, "Folder belongs to another user")
	}

	if err := fh.folders.Delete(folder.Uuid); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
