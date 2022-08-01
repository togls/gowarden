package raw

import (
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"

	"github.com/togls/gowarden/model"
	"github.com/togls/gowarden/store"
)

type folderStore struct {
	db *sql.DB
}

var _ store.Folder = (*folderStore)(nil)

func NewFolderStore(db *sql.DB) store.Folder {
	return &folderStore{db: db}
}

func (fs folderStore) Create(folder *model.Folder) error {
	now := time.Now()
	sql, args, err := squirrel.Insert("folders").
		Columns(fs.fields()...).
		Values(
			folder.Uuid,
			now,
			now,
			folder.UserUuid,
			folder.Name,
		).ToSql()
	if err != nil {
		return err
	}

	_, err = fs.db.Exec(sql, args...)
	return err
}

func (fs folderStore) FindByUuid(uuid string) (*model.Folder, error) {
	sqls, args, err := squirrel.Select(fs.fields()...).
		From("folders").
		Where(squirrel.Eq{"uuid": uuid}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var folder model.Folder
	err = fs.db.QueryRow(sqls, args...).Scan(
		&folder.Uuid,
		&folder.CreatedAt,
		&folder.UpdatedAt,
		&folder.UserUuid,
		&folder.Name,
	)

	if err == nil {
		return &folder, nil
	}

	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}

	return nil, err
}

func (fs folderStore) FindByUser(uuid string) ([]*model.Folder, error) {
	sqls, args, err := squirrel.Select(fs.fields()...).
		From("folders").
		Where(squirrel.Eq{"user_uuid": uuid}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := fs.db.Query(sqls, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	folders := []*model.Folder{}
	for rows.Next() {
		var folder model.Folder
		err = rows.Scan(
			&folder.Uuid,
			&folder.CreatedAt,
			&folder.UpdatedAt,
			&folder.UserUuid,
			&folder.Name,
		)
		if err != nil {
			return nil, err
		}

		folders = append(folders, &folder)
	}

	return folders, nil
}

func (fs folderStore) FindByUserCipher(user, cipher string) (*model.Folder, error) {
	sqls, args, err := squirrel.Select(fs.fields("f.")...).
		From("folders AS f").
		LeftJoin("folders_ciphers AS fc ON fc.folder_uuid = f.uuid").
		Where(squirrel.Eq{
			"f.user_uuid":    user,
			"fc.cipher_uuid": cipher,
		}).ToSql()
	if err != nil {
		return nil, err
	}

	var folder model.Folder
	err = fs.db.QueryRow(sqls, args...).Scan(
		&folder.Uuid,
		&folder.CreatedAt,
		&folder.UpdatedAt,
		&folder.UserUuid,
		&folder.Name,
	)

	if err == nil {
		return &folder, nil
	}

	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}

	return nil, err
}

func (fs folderStore) AddCipher(folder, cipher string) error {
	sql, args, err := squirrel.Insert("folders_ciphers").
		Columns("folder_uuid", "cipher_uuid").
		Values(folder, cipher).ToSql()
	if err != nil {
		return err
	}

	_, err = fs.db.Exec(sql, args...)
	return err
}

func (fs folderStore) Delete(uuid string) error {
	sqls, args, err := squirrel.Delete("folders").
		Where(squirrel.Eq{"uuid": uuid}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = fs.db.Exec(sqls, args...)
	return err
}

func (fs folderStore) DeleteAllByUser(user string) error {
	sql, args, err := squirrel.Delete("folders").
		Where(squirrel.Eq{"user_uuid": user}).ToSql()
	if err != nil {
		return err
	}

	_, err = fs.db.Exec(sql, args...)
	return err
}

func (fs folderStore) fields(prefix ...string) []string {

	ss := []string{
		"uuid",
		"created_at",
		"updated_at",
		"user_uuid",
		"name",
	}

	if len(prefix) != 1 {
		return ss
	}

	for i := range ss {
		ss[i] = prefix[0] + ss[i]
	}

	return ss
}
