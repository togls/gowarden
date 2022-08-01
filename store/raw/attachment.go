package raw

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/togls/gowarden/model"
	"github.com/togls/gowarden/store"
)

type attachmentStore struct {
	db *sql.DB
}

var _ store.Attachment = (*attachmentStore)(nil)

func NewAttachmentStore(db *sql.DB) store.Attachment {
	return &attachmentStore{db: db}
}

func (as attachmentStore) Find(cipher string) ([]*model.Attachment, error) {
	sql, args, err := squirrel.Select(as.fields()...).From("attachments").
		Where(squirrel.Eq{"cipher_uuid": cipher}).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := as.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []*model.Attachment
	for rows.Next() {
		attachment := new(model.Attachment)
		if err := rows.Scan(
			&attachment.ID,
			&attachment.CipherUuid,
			&attachment.FileName,
			&attachment.FileSize,
			&attachment.Akey,
		); err != nil {
			return nil, err
		}

		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

func (as attachmentStore) FindByUuid(uuid string) (*model.Attachment, error) {
	sql, args, err := squirrel.Select(as.fields()...).From("attachments").
		Where(squirrel.Eq{"id": uuid}).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := as.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, model.ErrNotFound
	}

	attachment := new(model.Attachment)
	if err := rows.Scan(
		&attachment.ID,
		&attachment.CipherUuid,
		&attachment.FileName,
		&attachment.FileSize,
		&attachment.Akey,
	); err != nil {
		return nil, err
	}

	return attachment, nil
}

func (as attachmentStore) fields() []string {
	return []string{
		"id",
		"cipher_uuid",
		"file_name",
		"file_size",
		"akey",
	}
}

func findAttachmentsByCipher(db *sql.DB, cipher string) ([]*model.Attachment, error) {
	panic("implement me")
}
