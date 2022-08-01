package raw

import (
	"database/sql"

	"github.com/Masterminds/squirrel"

	"github.com/togls/gowarden/model"
	"github.com/togls/gowarden/store"
)

type uoStore struct {
	db *sql.DB
}

var _ store.UserOrganization = (*uoStore)(nil)

func NewUserOrganizationStore(db *sql.DB) store.UserOrganization {
	return &uoStore{db: db}
}

func (uos uoStore) Find(filter *model.UOFilter) ([]*model.UserOrganization, error) {
	fields := []string{
		"uo.uuid",
		"uo.user_uuid",
		"uo.org_uuid",
		"uo.access_all",
		"uo.akey",
		"uo.status",
		"uo.atype",
		"o.name",
		"o.private_key",
		"o.public_key",
	}

	builder := squirrel.Select(fields...).From("users_organizations AS uo").
		LeftJoin("organizations AS o ON uo.org_uuid = o.uuid")

	if filter.UserUuid != nil {
		builder = builder.Where(squirrel.Eq{"user_uuid": *filter.UserUuid})
	}

	if filter.OrgUuid != nil {
		builder = builder.Where(squirrel.Eq{"org_uuid": *filter.OrgUuid})
	}

	if filter.Status != nil {
		builder = builder.Where(squirrel.Eq{"status": *filter.Status})
	}

	if filter.Atype != nil {
		builder = builder.Where(squirrel.Eq{"atype": *filter.Atype})
	}

	sqls, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := uos.db.Query(sqls, args...)
	if err != nil {
		return nil, err
	}

	var list []*model.UserOrganization
	for rows.Next() {
		var item model.UserOrganization
		err = rows.Scan(
			&item.Uuid,
			&item.UserUuid,
			&item.OrgUuid,
			&item.AccessAll,
			&item.AKey,
			&item.Status,
			&item.Atype,
			&item.Name,
			&item.PrivateKey,
			&item.PublicKey,
		)

		list = append(list, &item)
	}

	return list, nil
}

func (uos uoStore) FindByUuid(uuid string) (*model.UserOrganization, error) {
	fields := []string{
		"uo.uuid",
		"uo.user_uuid",
		"uo.org_uuid",
		"uo.access_all",
		"uo.akey",
		"uo.status",
		"uo.atype",
		"o.name",
		"o.private_key",
		"o.public_key",
	}
	sqls, args, err := squirrel.Select(fields...).From("users_organizations").
		LeftJoin("organizations AS o ON uo.org_uuid = o.uuid").
		Where(squirrel.Eq{"uuid": uuid}).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := uos.db.Query(sqls, args...)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, model.ErrNotFound
	}

	var item model.UserOrganization
	err = rows.Scan(
		&item.Uuid,
		&item.UserUuid,
		&item.OrgUuid,
		&item.AccessAll,
		&item.AKey,
		&item.Status,
		&item.Atype,
		&item.Name,
		&item.PrivateKey,
		&item.PublicKey,
	)
	return &item, err
}

func (uos uoStore) FindByUserAndOrg(user string, org string) (*model.UserOrganization, error) {
	fields := []string{
		"uo.uuid",
		"uo.user_uuid",
		"uo.org_uuid",
		"uo.access_all",
		"uo.akey",
		"uo.status",
		"uo.atype",
		"o.name",
		"o.private_key",
		"o.public_key",
	}
	sqls, args, err := squirrel.Select(fields...).From("users_organizations AS uo").
		LeftJoin("organizations AS o ON uo.org_uuid = o.uuid").
		Where(squirrel.Eq{"user_uuid": user, "org_uuid": org}).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := uos.db.Query(sqls, args...)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, model.ErrNotFound
	}

	var item model.UserOrganization
	err = rows.Scan(
		&item.Uuid,
		&item.UserUuid,
		&item.OrgUuid,
		&item.AccessAll,
		&item.AKey,
		&item.Status,
		&item.Atype,
		&item.Name,
		&item.PrivateKey,
		&item.PublicKey,
	)
	return &item, err
}

func (uos uoStore) Create(uo *model.UserOrganization) error {
	sqls, args, err := squirrel.Insert("users_organizations").
		Columns(uos.fields()...).
		Values(
			uo.Uuid,
			uo.UserUuid,
			uo.OrgUuid,
			uo.AccessAll,
			uo.AKey,
			uo.Status,
			uo.Atype,
		).ToSql()
	if err != nil {
		return err
	}

	_, err = uos.db.Exec(sqls, args...)
	return err
}

func (uos uoStore) Save(uo *model.UserOrganization) error {
	sqls, args, err := squirrel.Replace("users_organizations").
		Columns(uos.fields()...).
		Values(
			uo.Uuid,
			uo.UserUuid,
			uo.OrgUuid,
			uo.AccessAll,
			uo.AKey,
			uo.Status,
			uo.Atype,
		).ToSql()
	if err != nil {
		return err
	}

	_, err = uos.db.Exec(sqls, args...)
	return err
}

func (uos uoStore) Delete(uuid string) error {
	sqls, args, err := squirrel.Delete("users_organizations").
		Where(squirrel.Eq{"uuid": uuid}).ToSql()
	if err != nil {
		return err
	}

	_, err = uos.db.Exec(sqls, args...)
	return err
}

func (uos uoStore) DeleteAllByUser(user string) error {
	sqls, args, err := squirrel.Delete("users_organizations").
		Where(squirrel.Eq{"user_uuid": user}).ToSql()
	if err != nil {
		return err
	}

	_, err = uos.db.Exec(sqls, args...)
	return err
}

func (uos uoStore) fields() []string {
	return []string{
		"uuid",
		"user_uuid",
		"org_uuid",
		"access_all",
		"akey",
		"status",
		"atype",
	}
}
