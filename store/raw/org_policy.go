package raw

import (
	"database/sql"

	"github.com/Masterminds/squirrel"

	"github.com/togls/gowarden/model"
	"github.com/togls/gowarden/store"
)

type opStore struct {
	db *sql.DB
}

var _ store.OrgPolicy = (*opStore)(nil)

func NewOrgPolicyStore(db *sql.DB) store.OrgPolicy {
	return &opStore{db: db}
}

func (ops opStore) FindConfirmedByUser(user string) ([]*model.OrgPolicy, error) {
	sqls, args, err := squirrel.Select("op.*").From("org_policies AS op").
		InnerJoin("users_organizations AS uo ON uo.org_uuid = op.org_uuid").
		Where(squirrel.Eq{
			"uo.user_uuid": user,
			"uo.status":    model.UOStatusConfirmed,
		}).ToSql()
	if err != nil {
		return nil, err
	}

	raws, err := ops.db.Query(sqls, args...)
	if err != nil {
		return nil, err
	}

	var list []*model.OrgPolicy
	for raws.Next() {
		var op model.OrgPolicy
		err = raws.Scan(
			&op.Uuid,
			&op.OrgUuid,
			&op.Atype,
			&op.Enabled,
			&op.Data,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, &op)
	}

	return list, nil
}
