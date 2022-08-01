package raw

import (
	"database/sql"

	"github.com/Masterminds/squirrel"

	"github.com/togls/gowarden/model"
	"github.com/togls/gowarden/store"
)

type sendStore struct {
	db *sql.DB
}

var _ store.Send = (*sendStore)(nil)

func NewSendStore(db *sql.DB) store.Send {
	return &sendStore{db: db}
}

func (ss sendStore) Find(filter *model.SendFilter) ([]*model.Send, error) {
	builder := squirrel.Select("*").From("sends")

	if filter.UserUuid != nil {
		builder = builder.Where(
			squirrel.Eq{"user_uuid": *filter.UserUuid},
		)
	}

	sqls, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := ss.db.Query(sqls, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []*model.Send
	for rows.Next() {
		send, err := ss.scan(rows)
		if err != nil {
			return nil, err
		}

		list = append(list, send)
	}

	return list, nil
}

func (ss sendStore) DeleteAllByUser(userUuid string) error {
	_, err := ss.db.Exec("DELETE FROM sends WHERE user_uuid = ?", userUuid)
	return err
}

func (sendStore) fields() []string {
	return []string{
		"uuid",
		"user_uuid",
		"organization_uuid",
		"name",
		"notes",
		"atype",
		"data",
		"akey",
		"password_hash",
		"password_salt",
		"password_iter",
		"max_access_count",
		"access_count",
		"creation_date",
		"revision_date",
		"expiration_date",
		"deletion_date",
		"disabled",
		"hide_email",
	}
}

func (sendStore) scan(rows *sql.Rows) (*model.Send, error) {
	var send model.Send
	err := rows.Scan(
		&send.Uuid,
		&send.UserUuid,
		&send.OrganizationUuid,
		&send.Name,
		&send.Notes,
		&send.Atype,
		&send.Data,
		&send.Akey,
		&send.PasswordHash,
		&send.PasswordSalt,
		&send.PasswordIter,
		&send.MaxAccessCount,
		&send.AccessCount,
		&send.CreationDate,
		&send.RevisionDate,
		&send.ExpirationDate,
		&send.DeletionDate,
		&send.Disabled,
		&send.HideEmail,
	)

	return &send, err
}
