package store

import "github.com/togls/gowarden/model"

type User interface {
	Create(user *model.User) error
	Update(user *model.UpdateUser) error
	UpdateRevision(uuid string) error

	FindByEmail(string) (*model.User, error)
	FindByUuid(string) (*model.User, error)
}
