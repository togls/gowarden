package raw

import (
	"database/sql"

	"github.com/Masterminds/squirrel"

	"github.com/togls/gowarden/model"
	"github.com/togls/gowarden/store"
)

type emergencyAccessStore struct {
	db *sql.DB
}

var _ store.EmergencyAccess = (*emergencyAccessStore)(nil)

func NewEmergencyAccessStore(db *sql.DB) store.EmergencyAccess {
	return &emergencyAccessStore{db: db}
}

func (eas *emergencyAccessStore) Find(filter model.EAFilter) ([]*model.EmergencyAccess, error) {
	builder := squirrel.Select("*").From("emergency_access")

	if filter.GrantorUuid != nil {
		builder = builder.Where("grantor_uuid = ?", *filter.GrantorUuid)
	}

	if filter.GranteeUuid != nil {
		builder = builder.Where("grantee_uuid = ?", *filter.GranteeUuid)
	}

	panic("TODO: implement")
}

func (eas emergencyAccessStore) DeleteAllByUser(user string) error {
	sql, args, err := squirrel.Delete("emergency_access").
		Where(squirrel.Or{
			squirrel.Eq{"grantor_uuid": user},
			squirrel.Eq{"grantee_uuid": user},
		}).ToSql()
	if err != nil {
		return err
	}

	_, err = eas.db.Exec(sql, args...)
	return err
}
