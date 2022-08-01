package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/togls/gowarden/auth"
	"github.com/togls/gowarden/model"
)

// Share Cipher to Organization

type ShareCipherData struct {
	Cipher        CipherData `json:"Cipher"`
	CollectionIds []string   `json:"CollectionIds"`
}

func (ch *CipherHandler) PostCipherShare(c echo.Context) error {
	data := new(ShareCipherData)

	if err := c.Bind(data); err != nil {
		return err
	}

	cUuid := c.Param("uuid")

	user := auth.GetUser(c)

	cipher, err := ch.ciphers.FindByUuid(cUuid)
	if err != nil {
		return err
	}

	if err := ch.shareCipher(cipher, *data.Cipher.OrganizationId, user.Uuid, data.CollectionIds); err != nil {
		return err
	}

	return c.JSON(200, cipher)
}

func (ch *CipherHandler) PutCipherShare(c echo.Context) error {
	return ch.PostCipherShare(c)
}

type ShareSelectedCipherData struct {
	Ciphers       []CipherData `json:"Ciphers"`
	CollectionIds []string     `json:"CollectionIds"`
}

func (ch *CipherHandler) PutCipherShareSelected(c echo.Context) error {
	data := new(ShareSelectedCipherData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)

	for _, cipher := range data.Ciphers {
		oriCipher, err := ch.ciphers.FindByUuid(*cipher.ID)
		if err != nil {
			return err
		}

		if err := ch.shareCipher(oriCipher, *cipher.OrganizationId, user.Uuid, data.CollectionIds); err != nil {
			return err
		}
	}

	return c.NoContent(200)
}

func (ch *CipherHandler) shareCipher(cipher *model.Cipher, orgID, shareID string, collectionIDs []string) error {
	if cipher == nil || cipher.Uuid == "" || cipher.UserUuid == nil || *cipher.UserUuid != shareID {
		return echo.NewHTTPError(400, "Cipher not found")
	}

	if cipher.OrganizationUuid != nil {
		return echo.NewHTTPError(400, "Already belongs to an organization.")
	}

	// check cipher ownership
	ro, _, err := ch.accessRestrictions(cipher, shareID)
	if err != nil {
		return err
	}

	if ro {
		return echo.NewHTTPError(403, "Cipher is not write accessible")
	}

	// check collection access
	for _, collectionID := range collectionIDs {
		ok, err := ch.cs.CollectionWriteable(collectionID, shareID)
		if err != nil {
			return err
		}

		if !ok {
			return echo.NewHTTPError(403, "No rights to modify the collection")
		}
	}

	// save cipher
	cipher.OrganizationUuid = &orgID
	cipher.UserUuid = nil
	if err := ch.ciphers.Save(cipher); err != nil {
		return err
	}

	if err := ch.cs.SaveCipher(collectionIDs, cipher.Uuid); err != nil {
		return err
	}

	return nil
}
