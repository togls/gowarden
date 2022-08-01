package store

import "github.com/togls/gowarden/model"

type Cipher interface {
	Create(c *model.Cipher) error
	Save(c *model.Cipher) error

	FindByUuid(uuid string) (*model.Cipher, error)

	FindByOrg(org string) ([]*model.Cipher, error)
	FindByUser(uuid string) ([]*model.Cipher, error)
	FindByUserVisible(uuid string) ([]*model.Cipher, error)

	Delete(uuid string) error
	DeleteByOrg(org string) error
	DeleteByUser(user string) error
}
