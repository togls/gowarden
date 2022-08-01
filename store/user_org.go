package store

import "github.com/togls/gowarden/model"

type UserOrganization interface {
	Find(*model.UOFilter) ([]*model.UserOrganization, error)
	FindByUuid(uuid string) (*model.UserOrganization, error)
	FindByUserAndOrg(user, org string) (*model.UserOrganization, error)

	Create(uo *model.UserOrganization) error
	Save(uo *model.UserOrganization) error
	Delete(uuid string) error
	DeleteAllByUser(user string) error
}
