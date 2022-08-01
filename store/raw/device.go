package raw

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/togls/gowarden/model"
	"github.com/togls/gowarden/store"
)

type deviceStore struct {
	db *sql.DB
}

var _ store.Device = (*deviceStore)(nil)

func NewDeviceStore(db *sql.DB) store.Device {
	return newDeviceStore(db)
}

func newDeviceStore(db *sql.DB) *deviceStore {
	return &deviceStore{db: db}
}

func (ds deviceStore) Create(device *model.Device) error {
	now := time.Now()

	sql, args, err := squirrel.Insert("devices").
		Columns(ds.fields()...).
		Values(
			device.Uuid,
			now,
			now,
			device.UserUuid,
			device.Name,
			device.Atype,
			device.PushToken,
			device.RefreshToken,
			device.TwofactorRemember,
		).ToSql()
	if err != nil {
		return err
	}

	_, err = ds.db.Exec(sql, args...)
	return err
}

func (ds deviceStore) Save(device *model.Device) error {
	sql, args, err := squirrel.Replace("devices").
		Columns(ds.fields()...).
		Values(
			device.Uuid,
			device.CreatedAt,
			device.UpdatedAt,
			device.UserUuid,
			device.Name,
			device.Atype,
			device.PushToken,
			device.RefreshToken,
			device.TwofactorRemember,
		).ToSql()
	if err != nil {
		return err
	}

	_, err = ds.db.Exec(sql, args...)
	return err
}

func (ds deviceStore) Delete(uuid string) error {
	sql, args, err := squirrel.Delete("devices").
		Where(squirrel.Eq{"uuid": uuid}).ToSql()
	if err != nil {
		return err
	}

	_, err = ds.db.Exec(sql, args...)
	return err
}

func (ds deviceStore) FindByUuid(uuid string) (*model.Device, error) {
	sqls, args, err := squirrel.Select(ds.fields()...).From("devices").
		Where(squirrel.Eq{"uuid": uuid}).ToSql()
	if err != nil {
		return nil, err
	}

	return ds.findOne(sqls, args...)
}

func (ds deviceStore) FindByRefreshToken(token string) (*model.Device, error) {
	sqls, args, err := squirrel.Select(ds.fields()...).From("devices").
		Where(squirrel.Eq{"refresh_token": token}).
		ToSql()

	if err != nil {
		return nil, err
	}

	return ds.findOne(sqls, args...)
}

func (ds deviceStore) DeleteAllByUser(user string) error {
	sql, args, err := squirrel.Delete("devices").
		Where(squirrel.Eq{"user_uuid": user}).ToSql()
	if err != nil {
		return err
	}

	_, err = ds.db.Exec(sql, args...)
	return err
}

func (ds deviceStore) fields() []string {
	return []string{
		"uuid",
		"created_at",
		"updated_at",
		"user_uuid",
		"name",
		"atype",
		"push_token",
		"refresh_token",
		"twofactor_remember",
	}
}

func (ds deviceStore) findOne(sqls string, args ...any) (*model.Device, error) {
	var d model.Device
	err := ds.db.QueryRow(sqls, args...).Scan(
		&d.Uuid,
		&d.CreatedAt,
		&d.UpdatedAt,
		&d.UserUuid,
		&d.Name,
		&d.Atype,
		&d.PushToken,
		&d.RefreshToken,
		&d.TwofactorRemember,
	)

	if err == nil {
		return &d, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrNotFound
	}

	return nil, err
}
