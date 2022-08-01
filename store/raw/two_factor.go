package raw

import (
	"database/sql"

	"github.com/togls/gowarden/store"
)

type tfStore struct {
	db *sql.DB
}

var _ store.TwoFactor = (*tfStore)(nil)

func NewTwoFactorStore(db *sql.DB) store.TwoFactor {
	return &tfStore{db}
}

func (tfs tfStore) DeleteAllByUser(user string) error {
	_, err := tfs.db.Exec("DELETE FROM two_factor WHERE user_uuid = ?", user)
	return err
}

type tfiStore struct {
	db *sql.DB
}

var _ store.TwoFactorIncomplete = (*tfiStore)(nil)

func NewTwoFactorIncompleteStore(db *sql.DB) store.TwoFactorIncomplete {
	return &tfiStore{db: db}
}

func (tfis tfiStore) DeleteAllByUser(user string) error {
	_, err := tfis.db.Exec("DELETE FROM two_factor_incomplete WHERE user_uuid = ?", user)
	return err
}
