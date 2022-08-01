package raw

import (
	"database/sql"

	"github.com/Masterminds/squirrel"

	"github.com/togls/gowarden/model"
	"github.com/togls/gowarden/store"
)

type ucStore struct {
	db *sql.DB
}

var _ store.UserCollection = (*ucStore)(nil)

func NewUserCollectionStore(db *sql.DB) store.UserCollection {
	return newUCStore(db)
}

func newUCStore(db *sql.DB) *ucStore {
	return &ucStore{db}
}

func (ucs ucStore) Save(collection string, user string, readOnly bool, hidePasswords bool) error {
	sql, args, err := squirrel.Replace("users_collections").
		Columns(ucs.fields()...).
		Values(user, collection, readOnly, hidePasswords).
		ToSql()
	if err != nil {
		return err
	}

	result, err := ucs.db.Exec(sql, args...)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return model.ErrNotFound
	}

	return nil
}

func (ucs ucStore) Find(filter *model.UCFilter) (model.UCList, error) {
	builder := squirrel.Select(ucs.fields()...).
		From("users_collections")

	if filter.UserUuid != nil {
		builder = builder.Where(squirrel.Eq{"user_uuid": *filter.UserUuid})
	}

	if filter.OrgUuid != nil {
		builder = builder.Where(squirrel.Eq{"org_uuid": *filter.OrgUuid})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := ucs.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ucl model.UCList
	for rows.Next() {
		uc, err := ucs.scan(rows)
		if err != nil {
			return nil, err
		}
		ucl = append(ucl, uc)
	}

	return ucl, nil
}

func (ucs ucStore) FindByCollection(collection string) (model.UCList, error) {
	sql, args, err := squirrel.Select(ucs.fields()...).
		From("users_collections").
		Where(squirrel.Eq{"collection_uuid": collection}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := ucs.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ucl model.UCList
	for rows.Next() {
		uc, err := ucs.scan(rows)
		if err != nil {
			return nil, err
		}
		ucl = append(ucl, uc)
	}

	return ucl, nil
}

func (ucs ucStore) FindByCollectionUser(collection string, user string) (*model.UserCollection, error) {
	sql, args, err := squirrel.Select(ucs.fields()...).
		From("users_collections").
		Where(squirrel.Eq{
			"collection_uuid": collection,
			"user_uuid":       user,
		}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := ucs.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, model.ErrNotFound
	}

	uc, err := ucs.scan(rows)
	if err != nil {
		return nil, err
	}

	return uc, nil
}

func (ucs ucStore) FindByUserCipher(user string, cipher string) (*model.UserCollection, error) {
	sql, args, err := squirrel.Select("uc.*").From("ciphers AS c").
		InnerJoin("ciphers_collections AS cc ON cc.cipher_uuid = c.uuid").
		InnerJoin("users_collections AS uc ON uc.collection_uuid = cc.collection_uuid").
		Where(squirrel.Eq{
			"uc.user_uuid": user,
			"c.uuid":       cipher,
		}).ToSql()
	if err != nil {
		return nil, err
	}

	return ucs.findOne(sql, args...)
}

func (ucs ucStore) DeleteAllByCollection(collection string) error {
	sql, args, err := squirrel.Delete("users_collections").
		Where(squirrel.Eq{"collection_uuid": collection}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = ucs.db.Exec(sql, args...)
	return err
}

func (ucs ucStore) DeleteAllByUserAndOrg(user string, org string) error {
	const ucDeleteAllByUserAndOrg = `DELETE users_collections
	FROM
	  users_collections
	  INNER JOIN collections ON collections.uuid = users_collections.collection_uuid
	WHERE
	  users_collections.user_uuid = ?
	  AND collections.org_uuid = ?`

	args := []any{user, org}

	_, err := ucs.db.Exec(ucDeleteAllByUserAndOrg, args...)
	return err
}

func (ucs ucStore) DeleteByUserCollection(collection string, user string) error {
	const ucDeleteByUserCollection = `DELETE users_collections
	FROM
	  users_collections
	WHERE
	  user_uuid = ?
	  AND collection_uuid = ?`

	args := []any{user, collection}

	_, err := ucs.db.Exec(ucDeleteByUserCollection, args...)
	return err
}

func (ucStore) fields() []string {
	return []string{
		"user_uuid",
		"collection_uuid",
		"read_only",
		"hide_passwords",
	}
}

func (ucStore) scan(rows *sql.Rows) (*model.UserCollection, error) {
	uc := new(model.UserCollection)

	err := rows.Scan(
		&uc.UserUuid,
		&uc.CollectionUuid,
		&uc.ReadOnly,
		&uc.HidePasswords,
	)

	return uc, err
}

func (ucs ucStore) findOne(sql string, args ...any) (*model.UserCollection, error) {
	rows, err := ucs.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, model.ErrNotFound
	}

	uc, err := ucs.scan(rows)
	if err != nil {
		return nil, err
	}

	return uc, nil
}
