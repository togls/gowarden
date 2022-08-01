package model

import (
	"fmt"
)

type Attachment struct {
	ID         string
	CipherUuid string
	FileName   string
	FileSize   int
	Akey       *string
}

func (a Attachment) ToJson(host string) any {
	return struct {
		ID       string  `json:"Id"`
		Url      string  `json:"Url"`
		FileName string  `json:"FileName"`
		Size     string  `json:"Size"`
		SizeName string  `json:"SizeName"`
		Key      *string `json:"Key"`
		Object   string  `json:"Object"`
	}{
		ID:       a.ID,
		Url:      a.getUrl(host),
		FileName: a.FileName,
		Size:     fmt.Sprintf("%d", a.FileSize),
		SizeName: fmt.Sprintf("%d bytes", a.FileSize), // TODO: use humanize
		Key:      a.Akey,
		Object:   "attachment",
	}
}

func (a Attachment) getUrl(host string) string {
	return fmt.Sprintf("%s/attachments/%s/%s", host, a.CipherUuid, a.ID)
}
