package raw

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/togls/gowarden/model"
	"github.com/togls/gowarden/store"
)

type invitationStore struct {
	db *sql.DB
}

var _ store.Invitation = (*invitationStore)(nil)

func NewInvitationStore(db *sql.DB) store.Invitation {
	return &invitationStore{db: db}
}

func (is invitationStore) Save(invitation *model.Invitation) error {
	sqls, args, err := squirrel.Insert("invitations").
		Columns("email").
		Values(invitation.Email).
		ToSql()
	if err != nil {
		return err
	}

	_, err = is.db.Exec(sqls, args...)
	return err
}

func (is invitationStore) FindByEmail(email string) (*model.Invitation, error) {
	sqls, args, err := squirrel.Select("email").
		From("invitations").
		Where(squirrel.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var invitation model.Invitation
	err = is.db.QueryRow(sqls, args...).Scan(&invitation.Email)
	if err == nil {
		return &invitation, nil
	}

	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}

	return nil, err
}

func (is invitationStore) Delete(email string) error {
	sqls, args, err := squirrel.Delete("invitations").
		Where(squirrel.Eq{"email": email}).ToSql()
	if err != nil {
		return err
	}

	_, err = is.db.Exec(sqls, args...)
	return err
}
