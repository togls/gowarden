package raw

import (
	"database/sql"

	"github.com/Masterminds/squirrel"

	"github.com/togls/gowarden/model"
	"github.com/togls/gowarden/store"
)

type organizationStore struct {
	db *sql.DB
}

var _ store.Organization = (*organizationStore)(nil)

func NewOrganizationStore(db *sql.DB) store.Organization {
	return &organizationStore{db: db}
}

func (os organizationStore) FindByUuid(uuid string) (*model.Organization, error) {
	sqls, args, err := squirrel.Select(os.fields()...).From("organizations").
		Where(squirrel.Eq{"uuid": uuid}).ToSql()
	if err != nil {
		return nil, err
	}

	var item model.Organization
	err = os.db.QueryRow(sqls, args...).Scan(
		&item.Uuid,
		&item.Name,
		&item.BillingEmail,
		&item.PrivateKey,
		&item.PublicKey,
	)
	if err == nil {
		return &item, nil
	}

	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}

	return nil, err
}

func (os organizationStore) Create(org *model.Organization) error {
	sqls, args, err := squirrel.Insert("organizations").
		Columns(os.fields()...).
		Values(
			org.Uuid,
			org.Name,
			org.BillingEmail,
			org.PrivateKey,
			org.PublicKey,
		).ToSql()
	if err != nil {
		return err
	}

	_, err = os.db.Exec(sqls, args...)
	return err
}

func (os organizationStore) Save(org *model.Organization) error {
	sqls, args, err := squirrel.Replace("organizations").
		Columns(os.fields()...).
		Values(
			org.Uuid,
			org.Name,
			org.BillingEmail,
			org.PrivateKey,
			org.PublicKey,
		).ToSql()
	if err != nil {
		return err
	}

	_, err = os.db.Exec(sqls, args...)
	return err
}

func (os organizationStore) Delete(uuid string) error {
	sqls, args, err := squirrel.Delete("organizations").
		Where(squirrel.Eq{"uuid": uuid}).ToSql()
	if err != nil {
		return err
	}

	_, err = os.db.Exec(sqls, args...)
	return err
}

func (os organizationStore) fields() []string {
	return []string{
		"uuid",
		"name",
		"billing_email",
		"private_key",
		"public_key",
	}
}
