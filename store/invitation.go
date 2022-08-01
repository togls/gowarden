package store

import "github.com/togls/gowarden/model"

type Invitation interface {
	Save(invitation *model.Invitation) error
	FindByEmail(email string) (*model.Invitation, error)
	Delete(email string) error
}
