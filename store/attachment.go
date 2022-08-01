package store

import "github.com/togls/gowarden/model"

type Attachment interface {
	Find(cipher string) ([]*model.Attachment, error)

	FindByUuid(uuid string) (*model.Attachment, error)
}
