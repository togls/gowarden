package store

import "github.com/togls/gowarden/model"

type OrgPolicy interface {
	FindConfirmedByUser(userUUID string) ([]*model.OrgPolicy, error)
}
