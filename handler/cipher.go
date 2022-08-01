package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/togls/gowarden/auth"
	"github.com/togls/gowarden/config"
	"github.com/togls/gowarden/handler/response"
	"github.com/togls/gowarden/model"
	"github.com/togls/gowarden/pkg/crypto"
	"github.com/togls/gowarden/store"
)

type CipherHandler struct {
	logger *zerolog.Logger

	ciphers store.Cipher
	users   store.User
	folders store.Folder
	ucs     store.UserCollection
	cs      store.Collection
	uos     store.UserOrganization
	as      store.Attachment
	ops     store.OrgPolicy
	sends   store.Send
	favs    store.Favorite

	auth    *auth.Core
	globals config.GlobalDomains

	mailEnabled bool
}

func NewCipherHandler(
	logger *zerolog.Logger,
	globals config.GlobalDomains,
	auth *auth.Core,
	as store.Attachment,
	ciphers store.Cipher,
	cs store.Collection,
	favs store.Favorite,
	folders store.Folder,
	ops store.OrgPolicy,
	sends store.Send,
	ucs store.UserCollection,
	users store.User,
	uos store.UserOrganization,
) *CipherHandler {
	return &CipherHandler{
		logger: logger,

		ciphers: ciphers,
		users:   users,
		folders: folders,
		ucs:     ucs,
		cs:      cs,
		uos:     uos,
		as:      as,
		ops:     ops,
		sends:   sends,
		favs:    favs,

		auth:        auth,
		globals:     globals,
		mailEnabled: false,
	}
}

func (ch *CipherHandler) Routes(e *echo.Echo) {
	api := e.Group("/api", ch.auth.RequireAuth)
	api.GET("/sync", ch.Sync)

	cipher := api.Group("/ciphers")
	cipher.GET("", ch.GetCiphers)
	cipher.GET("/:uuid", ch.GetCipher)
	cipher.GET("/:uuid/admin", ch.GetCipherAdmin)
	cipher.GET("/:uuid/details", ch.GetCipherDetails)
	cipher.POST("", ch.PostCiphers)
	cipher.PUT("/:uuid/admin", ch.PutCipherAdmin)
	cipher.POST("/admin", ch.PostCiphersAdmin)
	cipher.POST("/create", ch.PostCiphersCreate)
	cipher.POST("/import", ch.PostCiphersImport)

	cipher.GET("/:uuid/attachment/:attachment_uuid", ch.GetAttachment)
	cipher.POST("/:uuid/attachment/v2", ch.PostAttachmentV2)
	cipher.POST("/:uuid/attachment/:attachment", ch.PostAttachmentV2Data)
	cipher.POST("/:uuid/attachment/:attachment", ch.PostAttachment)
	cipher.POST("/:uuid/attachment-admin", ch.PostAttachmentAdmin)
	cipher.POST("/:uuid/attachment/:attachment/share", ch.PostAttachmentShare)
	cipher.DELETE("/:uuid/attachment/:attachment", ch.DeleteAttachment)
	cipher.POST("/:uuid/attachment/:attachment/delete-admin", ch.DeleteAttachmentPostAdmin)
	cipher.DELETE("/:uuid/attachment/:attachment", ch.DeleteAttachment)
	cipher.DELETE("/:uuid/attachment/:attachment/admin", ch.DeleteAttachmentAdmin)

	cipher.POST("/:uuid/admin", ch.PostCipherAdmin)
	cipher.POST("/:uuid/share", ch.PostCipherShare)
	cipher.PUT("/:uuid/share", ch.PutCipherShare)
	cipher.PUT("/share", ch.PutCipherShareSelected)
	cipher.POST("/:uuid", ch.PostCipher)
	cipher.PUT("/:uuid", ch.PutCipher)

	cipher.POST("/:uuid/delete", ch.DeleteCipherPost)
	cipher.DELETE("/:uuid/delete-admin", ch.DeleteCipherAdmin)
	cipher.PUT("/:uuid/delete", ch.DeleteCipherPut)
	cipher.DELETE("/:uuid/delete-admin", ch.DeleteCipherPutAdmin)
	cipher.DELETE("/:uuid", ch.DeleteCipher)
	cipher.DELETE("/:uuid/admin", ch.DeleteCipherAdmin)
	cipher.DELETE("", ch.DeleteCipherSelected)
	cipher.POST("/delete", ch.DeleteCipherSelectedPost)
	cipher.PUT("/delete", ch.DeleteCipherSelectedPut)
	cipher.DELETE("/admin", ch.DeleteCipherSelectedAdmin)
	cipher.POST("/delete-admin", ch.DeleteCipherSelectedPostAdmin)
	cipher.PUT("/delete-admin", ch.DeleteCipherSelectedPutAdmin)
	cipher.PUT("/:uuid/restore", ch.RestoreCipherPut)
	cipher.PUT("/:uuid/restore-admin", ch.RestoreCipherPutAdmin)
	cipher.PUT("/restore", ch.RestoreCipherSelected)
	cipher.POST("/purge", ch.DeleteCipherAll)
	cipher.POST("/move", ch.MoveCipherSelected)
	cipher.PUT("/move", ch.MoveCipherSelectedPut)
	cipher.PUT("/:uuid/collections", ch.PutCollectionsUpdate)
	cipher.POST("/:uuid/collections", ch.PostCollectionsUpdate)
	cipher.POST("/:uuid/collections-admin", ch.PostCollectionsAdmin)
	cipher.PUT("/:uuid/collections-admin", ch.PutCollectionsAdmin)
}

type SyncData struct {
	ExcludeDomains bool `query:"excludeDomains"`
}

func (ch *CipherHandler) Sync(c echo.Context) error {
	data := new(SyncData)

	if err := c.Bind(data); err != nil {
		ch.logger.Debug().Err(err).Msg("Sync: failed to bind data")
		return err
	}

	user := auth.GetUser(c)

	uos, err := ch.uos.Find(&model.UOFilter{UserUuid: &user.Uuid})
	if err != nil {
		ch.logger.Debug().Err(err).Msg("Sync: failed to find user organizations")
		return err
	}

	folders, err := ch.folders.FindByUser(user.Uuid)
	if err != nil {
		ch.logger.Debug().Err(err).Msg("Sync: failed to find folders")
		return err
	}

	cs, err := ch.cs.Find(&model.CollectionFilter{UserUuid: &user.Uuid})
	if err != nil {
		ch.logger.Debug().Err(err).Msg("Sync: failed to find collections")
		return err
	}

	policies, err := ch.ops.FindConfirmedByUser(user.Uuid)
	if err != nil {
		ch.logger.Debug().Err(err).Msg("Sync: failed to find policies")
		return err
	}

	sends, err := ch.sends.Find(&model.SendFilter{UserUuid: &user.Uuid})
	if err != nil {
		ch.logger.Debug().Err(err).Msg("failed to find sends")
		return err
	}

	sendsData, err := response.NewSends(sends)
	if err != nil {
		ch.logger.Debug().Err(err).Msg("failed to convert sends data")
		return err
	}

	domains := new(response.Domains)
	if !data.ExcludeDomains {
		domains, err = getEqDomains(ch.globals, user, true)
		if err != nil {
			ch.logger.Debug().Err(err).Msg("Sync: failed to get domains")
			return err
		}
	}

	ciphers, err := ch.ciphers.FindByUserVisible(user.Uuid)
	if err != nil {
		ch.logger.Debug().Err(err).Msg("Sync: failed to find ciphers")
		return err
	}

	ciphersData := make([]*response.Cipher, 0, len(ciphers))
	for _, c := range ciphers {
		options := make([]response.CipherOption, 0)

		as, err := ch.as.Find(c.Uuid)
		if err != nil {
			ch.logger.Debug().Err(err).Str("cipher uuid", c.Uuid).Msg("")
			return err
		}

		options = append(options, response.CipherWithAttachments(as, "TODO: host"))

		f, err := ch.folders.FindByUserCipher(user.Uuid, c.Uuid)
		if err != nil && err != model.ErrNotFound {
			ch.logger.Debug().Err(err).Str("cipher uuid", c.Uuid).Msg("")
			return err
		} else if err == nil {
			options = append(options, response.CipherWithFolderId(f.Uuid))
		}

		ro, hp, err := ch.accessRestrictions(c, user.Uuid)
		if err != nil {
			ch.logger.Debug().Err(err).Str("cipher uuid", c.Uuid).Msg("")
			return err
		}

		options = append(options, response.CipherWithAccess(ro, hp))

		fav, err := ch.favs.IsFavorite(user.Uuid, c.Uuid)
		if err != nil {
			ch.logger.Debug().Err(err).Str("cipher uuid", c.Uuid).Msg("")
			return err
		}

		options = append(options, response.CipherWithFavorite(fav))

		cID, err := ch.cs.FindCollectionIds(c.Uuid, user.Uuid)
		if err != nil {
			ch.logger.Debug().Err(err).Str("cipher uuid", c.Uuid).Msg("")
			return err
		}

		options = append(options, response.CipherWithCollectionIds(cID))

		item, err := response.NewCipher(c, options...)
		if err != nil {
			ch.logger.Debug().Err(err).Str("cipher uuid", c.Uuid).Msg("failed to convert cipher data")
			return err
		}
		ciphersData = append(ciphersData, item)
	}

	resp := response.NewSync(
		ciphersData,
		response.NewCollectionDetailsList(cs),
		domains,
		response.NewFolders(folders),
		response.NewPolicies(policies),
		response.NewProfile(user, uos, false, ch.mailEnabled),
		sendsData,
	)

	return c.JSON(http.StatusOK, resp)
}

func (ch *CipherHandler) GetCiphers(c echo.Context) error {
	user := auth.GetUser(c)

	ciphers, err := ch.ciphers.FindByUserVisible(user.Uuid)
	if err != nil {
		return err
	}

	resp := &struct {
		Data              []*model.Cipher `json:"data"`
		Object            string          `json:"Object"`
		ContinuationToken *string         `json:"ContinuationToken"`
	}{
		Data:              ciphers,
		Object:            "list",
		ContinuationToken: nil,
	}

	return c.JSON(200, resp)
}

type RespCipherDetails struct {
	Object       string    `json:"Object"`
	ID           string    `json:"Id"`
	Type         int       `json:"Type"`
	RevisionDate time.Time `json:"RevisionDate"`
	DeletedDate  time.Time `json:"DeletedDate"`
	FolderID     string    `json:"FolderId"`
	Favorite     bool      `json:"Favorite"`
	Reprompt     bool      `json:"Reprompt"`
}

func (ch *CipherHandler) GetCipher(c echo.Context) error {
	uuid := c.Param("uuid")
	if uuid == "" {
		return echo.NewHTTPError(400, "Cipher doesn't exist")
	}

	user := auth.GetUser(c)

	cipher, err := ch.ciphers.FindByUuid(uuid)
	if err != nil {
		return echo.NewHTTPError(400, "Cipher doesn't exist")
	}

	if cipher.UserUuid != nil &&
		*cipher.UserUuid == user.Uuid {
		return c.JSON(200, cipher)
	}

	ro, _, err := ch.accessRestrictions(cipher, user.Uuid)
	if err != nil {
		return err
	}

	if ro {
		return echo.NewHTTPError(403, "Cipher is not owned by user")
	}

	return c.JSON(200, cipher)
}

func (ch *CipherHandler) GetCipherDetails(c echo.Context) error {
	return ch.GetCipher(c)
}

func (ch *CipherHandler) PostCiphers(c echo.Context) error {
	data := new(CipherData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)
	if data.FolderId != nil && *data.FolderId != "" {
		f, err := ch.folders.FindByUuid(*data.FolderId)
		if err != nil || f.UserUuid != user.Uuid {
			return echo.NewHTTPError(400, "Folders doesn't exist")
		}
	}

	newCipher, err := data.toCipher()
	if err != nil {
		return err
	}

	newUuid, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	newCipher.Uuid = newUuid.String()
	newCipher.UserUuid = &user.Uuid

	if err := ch.ciphers.Create(newCipher); err != nil {
		return err
	}

	// response options
	ops := make([]response.CipherOption, 0)

	// TODO: org
	// TODO: attachments

	if data.Favorite != nil && *data.Favorite {
		if err := ch.favs.AddFavorite(user.Uuid, newCipher.Uuid); err != nil {
			ch.logger.Debug().Err(err).Str("cipher uuid", newCipher.Uuid).Msg("")
			return err
		}

		ops = append(ops, response.CipherWithFavorite(*data.Favorite))
	}

	if data.FolderId != nil && *data.FolderId != "" {
		if err := ch.folders.AddCipher(*data.FolderId, newCipher.Uuid); err != nil {
			ch.logger.Debug().Err(err).Str("folder", *data.FolderId).Str("cipher uuid", newCipher.Uuid).Msg("")
			return err
		}

		ops = append(ops, response.CipherWithFolderId(*data.FolderId))
	}

	ops = append(ops, response.CipherWithAccess(false, false))

	resp, err := response.NewCipher(newCipher, ops...)
	if err != nil {
		ch.logger.Debug().Err(err).Str("cipher uuid", newCipher.Uuid).Msg("failed to convert cipher data")
		return err
	}

	return c.JSON(200, resp)
}

type ImportData struct {
	Ciphers             []CipherData       `json:"Ciphers"`
	Folders             []FolderData       `json:"Folders"`
	FolderRelationships []RelationshipData `json:"FolderRelationships"`
}

type RelationshipData struct {
	Key   int // cipher index
	Value int // folder index
}

func (ch *CipherHandler) PostCiphersImport(c echo.Context) error {
	data := new(ImportData)
	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)

	// Create folders
	for _, fd := range data.Folders {
		f, err := fd.toFolder(user.Uuid)
		if err != nil {
			return err
		}

		if err := ch.folders.Create(f); err != nil {
			return err
		}
	}

	rs := make(map[int]int)
	for _, fr := range data.FolderRelationships {
		rs[fr.Key] = fr.Value
	}

	// Create ciphers
	for index, cd := range data.Ciphers {

		c, err := cd.toCipher()
		if err != nil {
			return err
		}

		c.UserUuid = &user.Uuid

		if err := ch.ciphers.Create(c); err != nil {
			ch.logger.Debug().Err(err).Str("cipher", c.Uuid).Msg("")
			return err
		}

		folderID := data.Folders[rs[index]].Name
		if folderID == "" {
			continue
		}

		if err := ch.folders.AddCipher(folderID, c.Uuid); err != nil {
			ch.logger.Debug().Err(err).Str("folder", folderID).Msg("")
			return err
		}
	}

	now := time.Now()
	uu := &model.UpdateUser{
		Uuid:      user.Uuid,
		UpdatedAt: &now,
	}

	if err := ch.users.Update(uu); err != nil {
		return err
	}

	return c.NoContent(200)
}

func (ch *CipherHandler) PutCipher(c echo.Context) error {
	data := new(CipherData)

	if err := c.Bind(data); err != nil {
		return err
	}

	cUuid := c.Param("uuid")

	user := auth.GetUser(c)

	cipher, err := ch.ciphers.FindByUuid(cUuid)
	if err != nil {
		return err
	}

	ro, _, err := ch.accessRestrictions(cipher, user.Uuid)
	if err != nil {
		return err
	}

	if ro {
		return echo.NewHTTPError(403, "Cipher is not write accessible")
	}

	uc, err := data.toCipher()
	if err != nil {
		return err
	}

	uc.Uuid = cUuid
	uc.CreatedAt = cipher.CreatedAt
	uc.UserUuid = cipher.UserUuid
	uc.OrganizationUuid = cipher.OrganizationUuid

	if err := ch.ciphers.Save(uc); err != nil {
		return err
	}

	return c.JSON(200, uc)
}

func (ch *CipherHandler) PostCipher(c echo.Context) error {
	return ch.PutCipher(c)
}

func (ch *CipherHandler) DeleteCipher(c echo.Context) error {
	cUuid := c.Param("uuid")

	user := auth.GetUser(c)

	if err := ch.deleteCipher(cUuid, user.Uuid, false); err != nil {
		return err
	}

	return c.NoContent(200)
}

func (ch *CipherHandler) deleteCipher(cUuid, uUuid string, soft bool) error {
	cipher, err := ch.ciphers.FindByUuid(cUuid)
	if err != nil {
		return err
	}

	ro, _, err := ch.accessRestrictions(cipher, uUuid)
	if err != nil {
		return err
	}

	if ro {
		return echo.NewHTTPError(403, "Cipher is not write accessible")
	}

	if soft {
		now := time.Now()
		cipher.DeletedAt = &now

		if err := ch.ciphers.Save(cipher); err != nil {
			return err
		}

		return nil
	}

	if err := ch.ciphers.Delete(cipher.Uuid); err != nil {
		return err
	}

	return nil
}

func (ch *CipherHandler) DeleteCipherPost(c echo.Context) error {
	return ch.DeleteCipher(c)
}

func (ch *CipherHandler) DeleteCipherPut(c echo.Context) error {
	cUuid := c.Param("uuid")

	user := auth.GetUser(c)

	if err := ch.deleteCipher(cUuid, user.Uuid, true); err != nil {
		return err
	}

	return c.NoContent(200)
}

type DeleteCiphersData struct {
	IDs []string `json:"Ids"`
}

func (ch *CipherHandler) DeleteCipherSelected(c echo.Context) error {
	// soft := false
	data := new(DeleteCiphersData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)

	if err := ch.deleteCiphers(data.IDs, user.Uuid, false); err != nil {
		return err
	}

	return c.NoContent(200)
}

func (ch *CipherHandler) deleteCiphers(IDs []string, uUuid string, soft bool) error {
	for _, cUuid := range IDs {
		if err := ch.deleteCipher(cUuid, uUuid, soft); err != nil {
			return err
		}
	}
	return nil
}

func (ch *CipherHandler) DeleteCipherSelectedPost(c echo.Context) error {
	return ch.DeleteCipherSelected(c)
}

func (ch *CipherHandler) DeleteCipherSelectedPut(c echo.Context) error {
	data := new(DeleteCiphersData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)

	if err := ch.deleteCiphers(data.IDs, user.Uuid, true); err != nil {
		return err
	}

	return c.NoContent(200)
}

func (ch *CipherHandler) DeleteCipherAll(c echo.Context) error {
	data := new(PasswordData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)

	u, err := ch.users.FindByUuid(user.Uuid)
	if err != nil {
		return err
	}

	ok := crypto.VerifyPassword(
		data.MasterPasswordHash,
		u.Salt,
		u.PasswordHash,
		u.PasswordIterations,
	)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid password")
	}

	oUuid := c.Param("organizationId")
	if oUuid == "" {
		// purging user ciphers

		// delete ciphers
		cs, err := ch.ciphers.FindByUser(user.Uuid)
		if err != nil {
			return err
		}

		for _, c := range cs {
			if err := ch.ciphers.Delete(c.Uuid); err != nil {
				return err
			}
		}

		// delete folders
		fs, err := ch.folders.FindByUser(user.Uuid)
		if err != nil {
			return err
		}

		for _, f := range fs {
			if err := ch.folders.Delete(f.Uuid); err != nil {
				return err
			}
		}

		// update user revision
		now := time.Now()
		uu := &model.UpdateUser{
			Uuid: u.Uuid,

			UpdatedAt: &now,
		}
		if err := ch.users.Update(uu); err != nil {
			return err
		}

		return c.NoContent(200)
	}

	// delete organization ciphers

	uo, err := ch.uos.FindByUserAndOrg(user.Uuid, oUuid)
	if err != nil {
		return err
	}

	if uo.Atype != model.UOTypeOwner {
		return echo.NewHTTPError(http.StatusUnauthorized, "You don't have permission to purge the organization vault")
	}

	if err := ch.ciphers.DeleteByOrg(uo.OrgUuid); err != nil {
		return err
	}

	return c.NoContent(200)
}

func (ch *CipherHandler) RestoreCipherPut(c echo.Context) error {
	cUuid := c.Param("uuid")

	user := auth.GetUser(c)

	cipher, err := ch.ciphers.FindByUuid(cUuid)
	if err != nil {
		return err
	}

	ro, _, err := ch.accessRestrictions(cipher, user.Uuid)
	if err != nil {
		return err
	}

	if ro {
		return echo.NewHTTPError(http.StatusUnauthorized, "Cipher can't be restored by user")
	}

	cipher.DeletedAt = nil

	if err := ch.ciphers.Save(cipher); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, cipher)
}

func (ch *CipherHandler) RestoreCipherSelected(c echo.Context) error {
	data := new(DeleteCiphersData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)

	var ciphers []*model.Cipher
	for _, cUuid := range data.IDs {
		cipher, err := ch.ciphers.FindByUuid(cUuid)
		if err != nil {
			return err
		}

		ro, _, err := ch.accessRestrictions(cipher, user.Uuid)
		if err != nil {
			return err
		}

		if ro {
			return echo.NewHTTPError(http.StatusUnauthorized, "Cipher can't be restored by user")
		}

		cipher.DeletedAt = nil

		if err := ch.ciphers.Save(cipher); err != nil {
			return err
		}

		ciphers = append(ciphers, cipher)
	}

	resp := struct {
		Ciphers           []*model.Cipher `json:"Data"`
		Object            string          `json:"Object"`
		ContinuationToken *string         `json:"ContinuationToken"`
	}{
		Ciphers:           ciphers,
		Object:            "list",
		ContinuationToken: nil,
	}

	return c.JSON(http.StatusOK, &resp)
}

type MoveCiphersData struct {
	FolderID string   `json:"FolderId"`
	IDs      []string `json:"Ids"`
}

func (ch *CipherHandler) MoveCipherSelected(c echo.Context) error {
	data := new(MoveCiphersData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)

	folder, err := ch.folders.FindByUuid(data.FolderID)
	if err != nil {
		return err
	}

	if folder.UserUuid != user.Uuid {
		return echo.NewHTTPError(http.StatusUnauthorized, "Folder is not owned by user")
	}

	for _, cUuid := range data.IDs {
		cipher, err := ch.ciphers.FindByUuid(cUuid)
		if err != nil {
			return err
		}

		ro, _, err := ch.accessRestrictions(cipher, user.Uuid)
		if err != nil {
			return err
		}

		if ro {
			return echo.NewHTTPError(http.StatusUnauthorized, "Cipher can't be moved by user")
		}

		// TODO: save folder
		// cipher.FolderUuid = folder.Uuid

		if err := ch.ciphers.Save(cipher); err != nil {
			return err
		}
	}

	return c.NoContent(200)
}

func (ch *CipherHandler) MoveCipherSelectedPut(c echo.Context) error {
	return ch.MoveCipherSelected(c)
}

type CollectionsData struct {
	CollectionIDs []string `json:"CollectionIds"`
}

func (ch *CipherHandler) PutCollectionsUpdate(c echo.Context) error {
	return ch.PutCollectionsAdmin(c)
}

func (ch *CipherHandler) PostCollectionsUpdate(c echo.Context) error {
	return ch.PostCollectionsAdmin(c)
}

func (ch *CipherHandler) accessRestrictions(cipher *model.Cipher, uUuid string) (bool, bool, error) {
	if cipher.UserUuid != nil &&
		*cipher.UserUuid == uUuid {
		return false, false, nil
	}

	if cipher.OrganizationUuid != nil {
		uo, err := ch.uos.FindByUserAndOrg(uUuid, *cipher.OrganizationUuid)
		if err != nil {
			return false, false, fmt.Errorf("can't find user organization: %w", err)
		}

		if (uo.AccessAll || uo.Atype >= model.UOTypeAdmin) && uo.Status == model.UOStatusConfirmed {
			return false, false, nil
		}
	}

	uc, err := ch.ucs.FindByUserCipher(uUuid, cipher.Uuid)
	if err != nil {
		return false, false, fmt.Errorf("can't find user collection: %w", err)
	}

	return uc.ReadOnly, uc.HidePasswords, nil
}
