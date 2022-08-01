package raw

import (
	"database/sql"

	"github.com/Masterminds/squirrel"

	"github.com/togls/gowarden/store"
)

type favoriteStore struct {
	db *sql.DB
}

var _ store.Favorite = (*favoriteStore)(nil)

func NewFavoriteStore(db *sql.DB) store.Favorite {
	return &favoriteStore{db: db}
}

func (fs favoriteStore) IsFavorite(cipher, user string) (bool, error) {
	sqls, args, err := squirrel.Select("1").
		From("favorites").
		Where(squirrel.Eq{"user_uuid": user, "cipher_uuid": cipher}).
		Limit(1).ToSql()
	if err != nil {
		return false, err
	}

	var exists int
	err = fs.db.QueryRow(sqls, args...).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (fs favoriteStore) AddFavorite(cipher, user string) error {
	sql, args, err := squirrel.Insert("favorites").
		Columns("user_uuid", "cipher_uuid").
		Values(user, cipher).ToSql()
	if err != nil {
		return err
	}

	_, err = fs.db.Exec(sql, args...)
	return err
}

func (fs favoriteStore) DeleteAllByUser(user string) error {
	sql, args, err := squirrel.Delete("favorites").
		Where(squirrel.Eq{"user_uuid": user}).ToSql()
	if err != nil {
		return err
	}

	_, err = fs.db.Exec(sql, args...)
	return err
}
