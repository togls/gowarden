package raw

import (
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/togls/gowarden/model"
	"github.com/togls/gowarden/store"
)

type userStore struct {
	db *sql.DB
}

var _ store.User = (*userStore)(nil)

func NewUserStore(db *sql.DB) store.User {
	return &userStore{db: db}
}

func (us userStore) Create(user *model.User) error {
	sqls, args, err := squirrel.Insert("users").
		Columns(us.fields()...).
		Values(
			user.Uuid,
			user.CreatedAt,
			user.UpdatedAt,
			user.Email,
			user.Name,
			user.PasswordHash,
			user.Salt,
			user.PasswordIterations,
			user.PasswordHint,
			user.Akey,
			user.PrivateKey,
			user.PublicKey,
			user.TotpSecret,
			user.TotpRecover,
			user.SecurityStamp,
			user.EquivalentDomains,
			user.ExcludedGlobals,
			user.ClientKdfType,
			user.ClientKdfIter,
			user.VerifiedAt,
			user.LastVerifyingAt,
			user.LoginVerifyCount,
			user.EmailNew,
			user.EmailNewToken,
			user.Enabled,
			user.StampException,
			user.ApiKey,
		).ToSql()
	if err != nil {
		return err
	}

	_, err = us.db.Exec(sqls, args...)
	return err
}

func (us userStore) Update(user *model.UpdateUser) error {
	builder := squirrel.Update("users").
		Set("uuid", user.Uuid)

	if user.Name != nil {
		builder = builder.Set("name", *user.Name)
	}

	if user.PasswordHash != nil {
		builder = builder.Set("password_hash", user.PasswordHash)
	}

	if user.PasswordHint != nil {
		builder = builder.Set("password_hint", *user.PasswordHint)
	}

	if user.Akey != nil {
		builder = builder.Set("akey", *user.Akey)
	}

	if user.PublicKey != nil {
		builder = builder.Set("public_key", *user.PublicKey)
	}

	if user.PrivateKey != nil {
		builder = builder.Set("private_key", *user.PrivateKey)
	}

	if user.ClientKdfType != nil {
		builder = builder.Set("client_kdf_type", *user.ClientKdfType)
	}

	if user.ClientKdfIter != nil {
		builder = builder.Set("client_kdf_iter", *user.ClientKdfIter)
	}

	if user.SecurityStamp != nil {
		builder = builder.Set("security_stamp", *user.SecurityStamp)
	}

	if user.StampException != nil {
		builder = builder.Set("stamp_exception", *user.StampException)
	}

	if user.EmailNew != nil {
		builder = builder.Set("email_new", *user.EmailNew)
	}

	if user.Email != nil {
		builder = builder.Set("email", *user.Email)
	}

	if user.EmailNewToken != nil {
		builder = builder.Set("email_new_token", *user.EmailNewToken)
	}

	if user.ApiKey != nil {
		builder = builder.Set("api_key", *user.ApiKey)
	}

	if user.VerifiedAt != nil {
		builder = builder.Set("verified_at", *user.VerifiedAt)
	}

	if user.LastVerifyingAt != nil {
		builder = builder.Set("last_verifying_at", *user.LastVerifyingAt)
	}

	if user.LoginVerifyCount != nil {
		builder = builder.Set("login_verify_count", *user.LoginVerifyCount)
	}

	if user.UpdatedAt != nil {
		builder = builder.Set("updated_at", *user.UpdatedAt)
	}

	sqls, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = us.db.Exec(sqls, args...)
	return err
}

func (us userStore) UpdateRevision(uuid string) error {
	sqls, args, err := squirrel.Update("users").
		Set("updated_at", time.Now()).
		ToSql()
	if err != nil {
		return err
	}

	_, err = us.db.Exec(sqls, args...)
	return err
}

func (us userStore) FindByEmail(email string) (*model.User, error) {
	sqls, args, err := squirrel.Select(us.fields()...).
		From("users").
		Where(squirrel.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, err
	}

	return us.findOne(sqls, args...)
}

func (us userStore) FindByUuid(uuid string) (*model.User, error) {
	sqls, args, err := squirrel.Select(us.fields()...).
		From("users").
		Where(squirrel.Eq{"uuid": uuid}).
		ToSql()
	if err != nil {
		return nil, err
	}

	return us.findOne(sqls, args...)
}

func (us userStore) findOne(sqls string, args ...any) (*model.User, error) {
	rows, err := us.db.Query(sqls, args...)
	if err != nil {
		return nil, nil
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, model.ErrNotFound
	}

	return us.scan(rows)
}

func (userStore) scan(rows *sql.Rows) (*model.User, error) {
	var item model.User
	err := rows.Scan(
		&item.Uuid,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.Email,
		&item.Name,
		&item.PasswordHash,
		&item.Salt,
		&item.PasswordIterations,
		&item.PasswordHint,
		&item.Akey,
		&item.PrivateKey,
		&item.PublicKey,
		&item.TotpSecret,
		&item.TotpRecover,
		&item.SecurityStamp,
		&item.EquivalentDomains,
		&item.ExcludedGlobals,
		&item.ClientKdfType,
		&item.ClientKdfIter,
		&item.VerifiedAt,
		&item.LastVerifyingAt,
		&item.LoginVerifyCount,
		&item.EmailNew,
		&item.EmailNewToken,
		&item.Enabled,
		&item.StampException,
		&item.ApiKey,
	)
	return &item, err
}

func (us userStore) fields() []string {
	return []string{
		"uuid",
		"created_at",
		"updated_at",
		"email",
		"name",
		"password_hash",
		"salt",
		"password_iterations",
		"password_hint",
		"akey",
		"private_key",
		"public_key",
		"totp_secret",
		"totp_recover",
		"security_stamp",
		"equivalent_domains",
		"excluded_globals",
		"client_kdf_type",
		"client_kdf_iter",
		"verified_at",
		"last_verifying_at",
		"login_verify_count",
		"email_new",
		"email_new_token",
		"enabled",
		"stamp_exception",
		"api_key",
	}
}
