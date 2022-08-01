package handler

import "github.com/labstack/echo/v4"

func (ch *CipherHandler) GetAttachment(c echo.Context) error {
	cUuid := c.Param("uuid")
	if cUuid == "" {
		return echo.NewHTTPError(400, "Cipher doesn't exist")
	}

	aUuid := c.Param("attachment")
	if aUuid == "" {
		return echo.NewHTTPError(400, "Attachment doesn't exist")
	}

	attachment, err := ch.as.FindByUuid(aUuid)
	if err != nil {
		return err
	}

	if attachment.CipherUuid != cUuid {
		return echo.NewHTTPError(400, "Attachment doesn't belong to cipher")
	}

	// TODO: add hostname to url
	host := ""

	return c.JSON(200, attachment.ToJson(host))
}

// post_attachment_v2

func (ch *CipherHandler) PostAttachmentV2(c echo.Context) error {
	panic("TODO: not implemented")
}

// post_attachment_v2_data

func (ch *CipherHandler) PostAttachmentV2Data(c echo.Context) error {
	panic("TODO: not implemented")
}

func (ch *CipherHandler) PostAttachment(c echo.Context) error {
	panic("TODO: not implemented")
}

func (ch *CipherHandler) PostAttachmentAdmin(c echo.Context) error {
	panic("TODO: not implemented")
}

func (ch *CipherHandler) PostAttachmentShare(c echo.Context) error {
	panic("TODO: not implemented")
}

func (ch *CipherHandler) DeleteAttachmentPost(c echo.Context) error {
	panic("TODO: not implemented")
}

func (ch *CipherHandler) DeleteAttachmentPostAdmin(c echo.Context) error {
	panic("TODO: not implemented")
}

func (ch *CipherHandler) DeleteAttachment(c echo.Context) error {
	panic("TODO: not implemented")
}

func (ch *CipherHandler) DeleteAttachmentAdmin(c echo.Context) error {
	panic("TODO: not implemented")
}
