package store

import (
	"github.com/togls/gowarden/model"
)

type Device interface {
	Create(device *model.Device) error
	Save(device *model.Device) error
	Delete(uuid string) error
	FindByUuid(uuid string) (*model.Device, error)
	FindByRefreshToken(token string) (*model.Device, error)
	DeleteAllByUser(user string) error
}
