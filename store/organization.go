package store

import "github.com/togls/gowarden/model"

type Organization interface {
	FindByUuid(uuid string) (*model.Organization, error)

	Create(org *model.Organization) error
	Save(org *model.Organization) error
	Delete(uuid string) error
}
