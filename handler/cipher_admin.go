package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/togls/gowarden/auth"
	"github.com/togls/gowarden/model"
)

// Organization Cipher curd

func (ch *CipherHandler) GetCipherAdmin(c echo.Context) error {
	return ch.GetCipher(c)
}

func (ch *CipherHandler) PutCipherAdmin(c echo.Context) error {
	cUuid := c.Param("uuid")
	if cUuid == "" {
		return echo.NewHTTPError(400, "Cipher doesn't exist")
	}

	data := new(CipherData)
	if err := c.Bind(data); err != nil {
		return err
	}

	// TODO: put_cipher()

	panic("TODO: not implemented")
}

func (ch *CipherHandler) PostCipherAdmin(c echo.Context) error {
	return ch.PostCipher(c)
}

// post_ciphers_create
// Called when creating a new org-owned cipher, or cloning a cipher (whether
// user- or org-owned). When cloning a cipher to a user-owned cipher,
// `organizationId` is null.

func (ch *CipherHandler) PostCiphersCreate(c echo.Context) error {
	return ch.PostCiphersAdmin(c)
}

func (ch *CipherHandler) PostCiphersAdmin(c echo.Context) error {
	data := new(ShareCipherData)

	if err := c.Bind(data); err != nil {
		return err
	}

	if data.Cipher.OrganizationId != nil && len(data.CollectionIds) == 0 {
		return echo.NewHTTPError(400, "You must select at least one collection.")
	}

	// TODO: check org policy

	cipher, err := data.Cipher.toCipher()
	if err != nil {
		return err
	}

	user := auth.GetUser(c)
	cipher.UserUuid = &user.Uuid

	if err := ch.ciphers.Create(cipher); err != nil {
		return err
	}

	ch.shareCipher(cipher, *data.Cipher.OrganizationId, *cipher.UserUuid, data.CollectionIds)

	return c.JSON(200, cipher)
}

func (ch *CipherHandler) DeleteCipherAdmin(c echo.Context) error {
	return ch.DeleteCipher(c)
}

func (ch *CipherHandler) DeleteCipherPostAdmin(c echo.Context) error {
	return ch.DeleteCipherPost(c)
}

func (ch *CipherHandler) DeleteCipherPutAdmin(c echo.Context) error {
	return ch.DeleteCipherPut(c)
}

func (ch *CipherHandler) DeleteCipherSelectedAdmin(c echo.Context) error {
	return ch.DeleteCipherSelected(c)
}

func (ch *CipherHandler) DeleteCipherSelectedPostAdmin(c echo.Context) error {
	return ch.DeleteCipherSelectedPost(c)
}

func (ch *CipherHandler) DeleteCipherSelectedPutAdmin(c echo.Context) error {
	return ch.DeleteCipherSelectedPut(c)
}

func (ch *CipherHandler) RestoreCipherPutAdmin(c echo.Context) error {
	return ch.RestoreCipherPut(c)
}

func (ch *CipherHandler) PostCollectionsAdmin(c echo.Context) error {
	data := new(CollectionsData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)
	cUuid := c.Param("uuid")

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

	cs, err := ch.cs.Find(&model.CollectionFilter{})
	if err != nil {
		return err
	}

	different := func(a, b []string) []string {
		set := map[string]struct{}{}
		for _, x := range b {
			set[x] = struct{}{}
		}

		var list []string
		for _, x := range a {
			if _, ok := set[x]; !ok {
				list = append(list, x)
			}
		}
		return list
	}

	crtIDs := cs.IDs()
	addList := different(data.CollectionIDs, crtIDs)
	delList := different(crtIDs, data.CollectionIDs)

	// check if user has access to the collections
	for _, id := range append(addList, delList...) {
		collect, err := ch.cs.FindByUuid(id)
		if err != nil {
			return echo.NewHTTPError(400, "Invalid collection ID provided").SetInternal(err)
		}

		ok, err := ch.cs.CollectionWriteable(collect.Uuid, user.Uuid)
		if err != nil {
			return err
		}

		if !ok {
			return echo.NewHTTPError(403, "No rights to modify the collection")
		}

	}

	if err := ch.cs.SaveCipher(addList, cUuid); err != nil {
		return err
	}

	if err := ch.cs.DeleteCipher(delList, cUuid); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (ch *CipherHandler) PutCollectionsAdmin(c echo.Context) error {
	return ch.PostCollectionsAdmin(c)
}
