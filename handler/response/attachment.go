package response

import (
	"fmt"

	"github.com/togls/gowarden/model"
)

type Attachment struct {
	ID       string  `json:"Id"`
	Url      string  `json:"Url"`
	FileName string  `json:"FileName"`
	Size     string  `json:"Size"`
	SizeName string  `json:"SizeName"`
	Key      *string `json:"Key"`
	Object   string  `json:"Object"`
}

func NewAttachment(attachment *model.Attachment, host string) *Attachment {
	return &Attachment{
		ID:       attachment.ID,
		Url:      fmt.Sprintf("%s/attachments/%s/%s", host, attachment.CipherUuid, attachment.ID),
		FileName: attachment.FileName,
		Size:     fmt.Sprintf("%d", attachment.FileSize),
		SizeName: fmt.Sprintf("%d bytes", attachment.FileSize), // TODO: use humanize
		Key:      attachment.Akey,
		Object:   "attachment",
	}
}

func NewAttachments(attachments []*model.Attachment, host string) []*Attachment {
	items := make([]*Attachment, len(attachments))

	for i := range attachments {
		items[i] = NewAttachment(attachments[i], host)
	}

	return items
}
