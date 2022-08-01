package store

import "github.com/togls/gowarden/model"

type Send interface {
	Find(*model.SendFilter) ([]*model.Send, error)
	DeleteAllByUser(userUuid string) error
}
