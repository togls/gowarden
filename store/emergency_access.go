package store

import "github.com/togls/gowarden/model"

type EmergencyAccess interface {
	Find(filter model.EAFilter) ([]*model.EmergencyAccess, error)
	DeleteAllByUser(user string) error
}
