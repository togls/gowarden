package raw

import (
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"

	"github.com/togls/gowarden/model"
	"github.com/togls/gowarden/store"
)

type cipherStore struct {
	db *sql.DB
}

var _ store.Cipher = (*cipherStore)(nil)

func NewCipherStore(db *sql.DB) store.Cipher {
	return newCipherStore(db)
}

func newCipherStore(db *sql.DB) *cipherStore {
	return &cipherStore{db: db}
}

func (cs cipherStore) fields(prefix ...string) []string {
	ss := []string{
		"uuid",
		"user_uuid",
		"organization_uuid",
		"atype",
		"name",
		"notes",
		"fields",
		"data",
		"password_history",
		"reprompt",
		"created_at",
		"updated_at",
		"deleted_at",
	}

	if len(prefix) != 1 {
		return ss
	}

	for i := range ss {
		ss[i] = prefix[0] + ss[i]
	}

	return ss
}

func (cs cipherStore) scan(rows *sql.Rows) (*model.Cipher, error) {
	cipher := new(model.Cipher)

	err := rows.Scan(
		&cipher.Uuid,
		&cipher.UserUuid,
		&cipher.OrganizationUuid,
		&cipher.Atype,
		&cipher.Name,
		&cipher.Notes,
		&cipher.Fields,
		&cipher.Data,
		&cipher.PasswordHistory,
		&cipher.Reprompt,
		&cipher.CreatedAt,
		&cipher.UpdatedAt,
		&cipher.DeletedAt)
	if err != nil {
		return nil, err
	}

	return cipher, nil
}

func (cs cipherStore) Create(c *model.Cipher) error {
	now := time.Now()

	_, err := squirrel.Insert("ciphers").
		Columns(cs.fields()...).
		Values(
			c.Uuid,
			c.UserUuid,
			c.OrganizationUuid,
			c.Atype,
			c.Name,
			c.Notes,
			c.Fields,
			c.Data,
			c.PasswordHistory,
			c.Reprompt,
			now,
			now,
			nil,
		).
		RunWith(cs.db).Exec()

	return err
}

func (cs cipherStore) Save(c *model.Cipher) error {
	result, err := squirrel.Replace("ciphers").Columns(cs.fields()...).Values(
		c.Uuid,
		c.UserUuid,
		c.OrganizationUuid,
		c.Atype,
		c.Name,
		c.Notes,
		c.Fields,
		c.Data,
		c.PasswordHistory,
		c.Reprompt,
		c.CreatedAt,
		c.UpdatedAt,
		c.DeletedAt,
	).RunWith(cs.db).Exec()
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return model.ErrNotFound
	}

	return nil
}

func (cs cipherStore) FindByUuid(uuid string) (*model.Cipher, error) {
	sql, args, err := squirrel.Select(cs.fields()...).From("ciphers").
		Where(squirrel.Eq{"uuid": uuid}).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := cs.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, model.ErrNotFound
	}

	cipher, err := cs.scan(rows)
	if err != nil {
		return nil, err
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cipher, nil
}

func (cs cipherStore) FindByOrg(org string) ([]*model.Cipher, error) {
	return cs.find(&cipherFilter{org: &org})
}

func (cs cipherStore) FindByUser(user string) ([]*model.Cipher, error) {
	return cs.find(&cipherFilter{user: &user})
}

func (cs cipherStore) FindByUserVisible(user string) ([]*model.Cipher, error) {
	fs := cs.fields("c.")

	sqls, args, err := squirrel.Select(fs...).From("ciphers AS c").
		LeftJoin("ciphers_collections AS cc ON cc.cipher_uuid = c.uuid").
		LeftJoin("users_organizations AS uo ON uo.org_uuid = c.organization_uuid").
		LeftJoin("users_collections AS uc ON uc.collection_uuid = cc.collection_uuid").
		Where(
			squirrel.Or{
				squirrel.Eq{"c.user_uuid": user},
				squirrel.And{ // user is in org and access all
					squirrel.Eq{"uo.user_uuid": user},
					squirrel.Eq{"uo.access_all": true},
				}, // or user access collection
				squirrel.Eq{"uc.user_uuid": user},
			},
		).Distinct().ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := cs.db.Query(sqls, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ciphers []*model.Cipher
	for rows.Next() {
		var cipher model.Cipher
		rows.Scan(
			&cipher.Uuid,
			&cipher.UserUuid,
			&cipher.OrganizationUuid,
			&cipher.Atype,
			&cipher.Name,
			&cipher.Notes,
			&cipher.Fields,
			&cipher.Data,
			&cipher.PasswordHistory,
			&cipher.Reprompt,
			&cipher.CreatedAt,
			&cipher.UpdatedAt,
			&cipher.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		ciphers = append(ciphers, &cipher)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ciphers, nil
}

func (cs cipherStore) Delete(uuid string) error {
	result, err := squirrel.Delete("ciphers").Where(squirrel.Eq{"uuid": uuid}).
		RunWith(cs.db).Exec()

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return model.ErrNotFound
	}

	return nil
}

func (cs cipherStore) DeleteByOrg(org string) error {
	result, err := squirrel.Delete("ciphers").Where(squirrel.Eq{"organization_uuid": org}).
		RunWith(cs.db).Exec()

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return model.ErrNotFound
	}

	return nil
}

func (cs cipherStore) DeleteByUser(user string) error {
	_, err := squirrel.Delete("ciphers").Where(squirrel.Eq{"user_uuid": user}).
		RunWith(cs.db).Exec()
	return err
}

type cipherFilter struct {
	org     *string
	user    *string
	visible *bool
}

func (cs cipherStore) find(filter *cipherFilter) ([]*model.Cipher, error) {
	var ciphers []*model.Cipher

	builder := squirrel.Select(cs.fields()...).From("ciphers")

	if filter.org != nil {
		builder = builder.Where(squirrel.Eq{"organization_uuid": *filter.org})
	}

	if filter.user != nil {
		builder = builder.Where(squirrel.Eq{"user_uuid": *filter.user})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := cs.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		cipher, err := cs.scan(rows)
		if err != nil {
			return nil, err
		}

		ciphers = append(ciphers, cipher)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ciphers, nil
}
