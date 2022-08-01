package store

import "github.com/togls/gowarden/model"

type Folder interface {
	Create(folder *model.Folder) error

	FindByUser(uuid string) ([]*model.Folder, error)

	FindByUuid(uuid string) (*model.Folder, error)
	FindByUserCipher(user, cipher string) (*model.Folder, error)

	AddCipher(folder, cipher string) error

	Delete(uuid string) error
	DeleteAllByUser(user string) error
}
