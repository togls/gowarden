package raw

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"

	"github.com/togls/gowarden/model"
	"github.com/togls/gowarden/store"
)

type collectionStore struct {
	db *sql.DB
}

var _ store.Collection = (*collectionStore)(nil)

func NewCollectionStore(db *sql.DB) store.Collection {
	return &collectionStore{db: db}
}

func (cstore collectionStore) Find(filter *model.CollectionFilter) (model.CollectionList, error) {
	builder := squirrel.Select(cstore.fields()...).From("collections")

	if filter.OrgUuid != nil {
		builder = builder.Where(squirrel.Eq{"org_uuid": *filter.OrgUuid})
	}

	if filter.UserUuid != nil {
		fs := cstore.fields()
		fs = append(fs, "uc.read_only", "uc.hide_passwords")
		builder = squirrel.Select(fs...).From("collections").
			LeftJoin("users_collections AS uc ON uc.collection_uuid = collections.uuid").
			Where(squirrel.Eq{"uc.user_uuid": *filter.UserUuid})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := cstore.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list model.CollectionList
	for rows.Next() {
		var cl model.Collection
		var err error
		if filter.UserUuid == nil {
			err = rows.Scan(&cl.Uuid, &cl.OrgUuid, &cl.Name)
		} else {
			err = rows.Scan(&cl.Uuid, &cl.OrgUuid, &cl.Name, &cl.ReadOnly, &cl.HidePasswords)
		}
		if err != nil {
			return nil, err
		}

		list = append(list, &cl)
	}

	return list, nil
}

func (cstore collectionStore) FindByUuid(uuid string) (*model.Collection, error) {
	sql, args, err := squirrel.Select(cstore.fields()...).From("collections").
		Where(squirrel.Eq{"uuid": uuid}).ToSql()
	if err != nil {
		return nil, err
	}

	return cstore.findOne(sql, args)
}

func (cstore collectionStore) FindByCipherAndOrg(cipher string, org string) (*model.Collection, error) {
	builder := squirrel.Select(cstore.fields()...).From("collections AS c").
		LeftJoin("ciphers_collections AS cc ON cc.collection_uuid = c.uuid").
		Where(squirrel.And{
			squirrel.Eq{"cc.cipher_uuid": cipher},
			squirrel.Eq{"c.org_uuid": org},
		})

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	return cstore.findOne(sql, args)
}

func (cstore collectionStore) FindByCollectionUser(collection string, user string) (*model.Collection, error) {
	builder := squirrel.Select(cstore.fields()...).From("collections AS c").
		LeftJoin("users_collections AS uc ON uc.collection_uuid = c.uuid").
		Where(squirrel.And{
			squirrel.Eq{"c.uuid": collection},
			squirrel.Eq{"uc.user_uuid": user},
		})

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	return cstore.findOne(sql, args)
}

func (cstore collectionStore) FindByCollectionOrg(collection string, org string) (*model.Collection, error) {
	builder := squirrel.Select(cstore.fields()...).From("collections").
		Where(squirrel.And{
			squirrel.Eq{"uuid": collection},
			squirrel.Eq{"org_uuid": org},
		})

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	return cstore.findOne(sql, args)
}

func (cstore collectionStore) Save(c *model.Collection) error {
	sql, args, err := squirrel.Replace("collections").
		Columns(cstore.fields()...).
		Values(c.Uuid, c.OrgUuid, c.Name).ToSql()
	if err != nil {
		return err
	}

	result, err := cstore.db.Exec(sql, args...)
	if err != nil {
		return err
	}

	if _, err := result.RowsAffected(); err != nil {
		return err
	}

	return nil
}

func (cstore collectionStore) Delete(uuid string) error {
	sql, args, err := squirrel.Delete("collections").
		Where(squirrel.Eq{"uuid": uuid}).ToSql()
	if err != nil {
		return err
	}

	result, err := cstore.db.Exec(sql, args...)
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

func (cstore collectionStore) FindCollectionIds(cipher, user string) ([]string, error) {
	builder := squirrel.Select("c.uuid").From("collections AS c").
		LeftJoin("ciphers_collections AS cc ON cc.collection_uuid = c.uuid").
		LeftJoin("users_collections AS uc ON uc.collection_uuid = c.uuid").
		LeftJoin("users_organizations AS uo ON uo.org_uuid = c.org_uuid").
		Where(squirrel.And{
			squirrel.Eq{"cc.cipher_uuid": cipher},
			squirrel.Eq{"uo.user_uuid": user},
			squirrel.Or{
				squirrel.Eq{"uc.user_uuid": user},
				squirrel.Eq{"uo.access_all": true},
				squirrel.LtOrEq{"uo.atype": int(model.UOTypeAdmin)},
			},
		}).Distinct()

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := cstore.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []string
	for rows.Next() {
		var uuid string
		if err := rows.Scan(&uuid); err != nil {
			return nil, err
		}
		list = append(list, uuid)
	}

	return list, nil
}

func (cstore collectionStore) SaveCipher(collectionIDs []string, cipher string) error {
	builder := squirrel.Insert("ciphers_collections").
		Columns("cipher_uuid", "collection_uuid")

	for _, collectionID := range collectionIDs {
		builder = builder.Values(cipher, collectionID)
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	result, err := cstore.db.Exec(sql, args...)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n != int64(len(collectionIDs)) {
		return errors.New("unexpected number of rows affected")
	}

	return nil
}

func (cstore collectionStore) DeleteCipher(collectionIDs []string, cipher string) error {
	var or squirrel.Or
	for _, collectionID := range collectionIDs {
		or = append(or, squirrel.Or{
			squirrel.Eq{"cipher_uuid": cipher},
			squirrel.Eq{"collection_uuid": collectionID},
		})
	}

	sql, args, err := squirrel.Delete("ciphers_collections").Where(or).ToSql()
	if err != nil {
		return err
	}

	result, err := cstore.db.Exec(sql, args...)
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

// UserCollection
func (cstore collectionStore) SaveUser(collectionIDs []string, user string, readOnly bool, hidePasswords bool) error {
	builder := squirrel.Insert("users_collections").
		Columns("user_uuid", "collection_uuid", "read_only", "hide_passwords")

	for _, collectionID := range collectionIDs {
		builder = builder.Values(user, collectionID, readOnly, hidePasswords)
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	result, err := cstore.db.Exec(sql, args...)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n != int64(len(collectionIDs)) {
		return errors.New("unexpected number of rows affected")
	}

	return nil
}

func (cstore collectionStore) DeleteUser(collectionIDs []string, user string) error {
	var or squirrel.Or
	for _, collectionID := range collectionIDs {
		or = append(or, squirrel.Or{
			squirrel.Eq{"user_uuid": user},
			squirrel.Eq{"collection_uuid": collectionID},
		})
	}

	sql, args, err := squirrel.Delete("users_collections").Where(or).ToSql()
	if err != nil {
		return err
	}

	result, err := cstore.db.Exec(sql, args...)
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

func (cstore collectionStore) CollectionWriteable(collection string, user string) (bool, error) {
	sqls, args, err := squirrel.Select("read_only").From("users_collections").
		Where(squirrel.And{
			squirrel.Eq{"collection_uuid": collection},
			squirrel.Eq{"user_uuid": user},
		}).ToSql()
	if err != nil {
		return false, err
	}

	row := cstore.db.QueryRow(sqls, args)

	var readOnly bool
	err = row.Scan(&readOnly)
	if err == nil {
		return !readOnly, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	return false, err
}

func (collectionStore) fields() []string {
	return []string{
		"uuid",
		"org_uuid",
		"name",
	}
}

func (cstore collectionStore) findOne(sql string, args any) (*model.Collection, error) {
	rows, err := cstore.db.Query(sql, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, model.ErrNotFound
	}

	return cstore.scan(rows)
}

func (collectionStore) scan(rows *sql.Rows) (*model.Collection, error) {
	cl := new(model.Collection)
	err := rows.Scan(&cl.Uuid, &cl.OrgUuid, &cl.Name)
	return cl, err
}

// $1=cipher_uuid, $2=user_uuid
var GetCollections = fmt.Sprintf(`
SELECT cc.collection_uuid FROM ciphers_collections AS cc
INNER JOIN collections AS c ON cc.collection_uuid = c.uuid
INNER JOIN users_organizations AS uo ON c.org_uuid = uo.org_uuid AND uo.user_uuid = $2
LEFT JOIN users_collections AS uc ON cc.collection_uuid = uc.collection_uuid AND uc.user_uuid = $2
WHERE cc.cipher_uuid = $1 AND uc.user_uuid = $2 
OR uo.access_all = true 
OR uo.atype = %d`, model.UOTypeAdmin)
