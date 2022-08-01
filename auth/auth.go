package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/wire"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/togls/gowarden/config"
	"github.com/togls/gowarden/model"
	"github.com/togls/gowarden/pkg/crypto"
	"github.com/togls/gowarden/store"
)

var WireSet = wire.NewSet(
	New,

	wire.Bind(new(Authenticator), new(*Core)),
	wire.Bind(new(JWTDecoder), new(*Core)),
)

type Authenticator interface {
	RefreshLogin(token string) (*RespRefreshToken, error)
	PasswordLogin(cd *ConnectData) (*RespRefreshToken, error)
}

type JWTDecoder interface {
	DecodeToken(token string, claims jwt.Claims) error
}

type Core struct {
	logger *zerolog.Logger

	devices store.Device
	users   store.User
	uos     store.UserOrganization
	ucs     store.UserCollection

	validity time.Duration

	sm jwt.SigningMethod

	mailEnabled bool
	priKey      *rsa.PrivateKey
	pubKey      *rsa.PublicKey
}

var _ Authenticator = (*Core)(nil)

var _ JWTDecoder = (*Core)(nil)

func New(
	cfg *config.Core,
	d store.Device,
	u store.User,
	uos store.UserOrganization,
	ucs store.UserCollection,
) *Core {
	return &Core{
		priKey:      cfg.PriKey,
		pubKey:      cfg.PubKey,
		mailEnabled: false,
		logger:      cfg.Logger,

		devices: d,
		users:   u,
		uos:     uos,
		ucs:     ucs,

		validity: time.Hour * 2,
		sm:       jwt.GetSigningMethod("RS256"),
	}
}

func (core Core) RefreshLogin(token string) (*RespRefreshToken, error) {
	d, err := core.devices.FindByRefreshToken(token)
	if err != nil {
		core.logger.Debug().Err(err).Msg("device not found")
		return nil, err
	}

	u, err := core.users.FindByUuid(d.UserUuid)
	if err != nil {
		core.logger.Debug().Err(err).Msg("user not found")
		return nil, err
	}

	accessToken, err := core.refreshToken(u, d)
	if err != nil {
		core.logger.Debug().Err(err).Msg("refresh token")
		return nil, err
	}

	err = core.devices.Save(d)
	if err != nil {
		core.logger.Debug().Err(err).Msg("save device")
		return nil, err
	}

	return &RespRefreshToken{
		AccessToken:  accessToken,
		ExpiresIn:    core.validity.Seconds(),
		TokenType:    "Bearer",
		RefreshToken: d.RefreshToken,
		Key:          u.Akey,
		PrivateKey:   u.PrivateKey,

		Kdf:                 u.ClientKdfType,
		KdfIterations:       u.ClientKdfIter,
		ResetMasterPassword: false,
		Scope:               "api offline_access",
		UnofficialServer:    true,
	}, nil
}

func (core Core) PasswordLogin(cd *ConnectData) (*RespRefreshToken, error) {
	if cd.Scope != "api offline_access" {
		return nil, ErrScopeNotSupported
	}

	u, err := core.users.FindByEmail(cd.Username)
	if err != nil {
		core.logger.Info().
			Err(err).
			Str("email", cd.Username).
			Msg("user not found")
		return nil, echo.NewHTTPError(http.StatusUnauthorized,
			"invalid username or password")
	}

	ok := crypto.VerifyPassword(
		cd.Password,
		u.Salt,
		u.PasswordHash,
		u.PasswordIterations,
	)
	if !ok {
		core.logger.Info().
			Str("password", cd.Password).
			Str("email", cd.Username).
			Msg("password mismatch")
		return nil, echo.NewHTTPError(http.StatusUnauthorized,
			"Username or password is incorrect. Try again")
	}

	if !u.Enabled {
		core.logger.Info().Str("email", cd.Username).Msg("user disabled")
		return nil, ErrUserDisabled
	}

	if core.mailEnabled && u.VerifiedAt == nil {
		core.logger.Info().Str("email", cd.Username).Msg("user not verified")
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "User is not verified")
	}

	d, err := core.devices.FindByUuid(cd.DeviceIdentifier)
	if err != nil || d.UserUuid != u.Uuid {
		// create new device
		t, _ := strconv.Atoi(cd.DeviceType)
		newDevice := &model.Device{
			Uuid:     cd.DeviceIdentifier,
			UserUuid: u.Uuid,
			Name:     cd.DeviceName,
			Atype:    t,

			CreatedAt: time.Now(),
		}

		d = newDevice

	}
	// TODO: send mail

	// TODO: twofactor_auth

	accessToken, err := core.refreshToken(u, d)
	if err != nil {
		core.logger.Debug().Err(err).Msg("refresh token")
		return nil, err
	}

	d.UpdatedAt = time.Now()
	if err := core.devices.Save(d); err != nil {
		core.logger.Debug().Err(err).Str("device uuid", d.Uuid).Msg("save device")
		return nil, err
	}

	return &RespRefreshToken{
		AccessToken:  accessToken,
		ExpiresIn:    core.validity.Seconds(),
		TokenType:    "Bearer",
		RefreshToken: d.RefreshToken,
		Key:          u.Akey,
		PrivateKey:   u.PrivateKey,

		Kdf:                 u.ClientKdfType,
		KdfIterations:       u.ClientKdfIter,
		ResetMasterPassword: false,
		Scope:               "api offline_access",
		UnofficialServer:    true,
	}, nil
}

func (core Core) refreshToken(u *model.User, d *model.Device) (string, error) {
	if d.RefreshToken == "" {
		src, err := crypto.GenerateBytes(64)
		if err != nil {
			return "", err
		}

		dst := make([]byte, base64.StdEncoding.EncodedLen(len(src)))
		base64.StdEncoding.Encode(dst, src)

		d.RefreshToken = string(dst)
	}

	confirmed := model.UOStatusConfirmed
	filter := &model.UOFilter{UserUuid: &u.Uuid, Status: &confirmed}
	userOrgs, err := core.uos.Find(filter)
	if err != nil {
		return "", err
	}

	orgowner := make([]string, 0)
	orgadmin := make([]string, 0)
	orguser := make([]string, 0)
	orgmanager := make([]string, 0)
	for _, uo := range userOrgs {
		switch uo.Atype {
		case model.UOTypeOwner:
			orgowner = append(orgowner, uo.OrgUuid)
		case model.UOTypeAdmin:
			orgadmin = append(orgadmin, uo.OrgUuid)
		case model.UOTypeUser:
			orguser = append(orguser, uo.OrgUuid)
		case model.UOTypeManager:
			orgmanager = append(orgmanager, uo.OrgUuid)
		}
	}

	now := time.Now()

	claims := &LoginJwtClaims{
		RegisteredClaims: &jwt.RegisteredClaims{
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(core.validity)),
			Issuer:    "|login",
			Subject:   u.Uuid,
		},

		Orgowner:   orgowner,
		Orgadmin:   orgadmin,
		Orguser:    orguser,
		Orgmanager: orgmanager,

		Device:        d.Uuid,
		Name:          u.Name,
		Email:         u.Email,
		EmailVerified: (!core.mailEnabled || u.VerifiedAt != nil),

		Sstamp:  u.SecurityStamp,
		Premium: true,
		Scope:   []string{"api", "offline_access"},
		Amr:     []string{"Application"},
	}

	if !core.mailEnabled || u.VerifiedAt != nil {
		claims.EmailVerified = true
	}

	t := jwt.New(core.sm)
	t.Claims = claims
	accessToken, err := t.SignedString(core.priKey)
	if err != nil {
		return "", err
	}

	d.UpdatedAt = now
	return accessToken, nil
}

func (core Core) DecodeToken(token string, claims jwt.Claims) error {
	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return core.pubKey, nil
	})
	if err != nil {
		return err
	}

	if !t.Valid {
		return errors.New("invalid claim")
	}

	return nil
}

type GrantType int

const (
	GTRefreshToken GrantType = iota
	GTPassword
	ClientCredentials
)

func (gt *GrantType) UnmarshalParam(param string) error {
	switch param {
	case "refresh_token":
		*gt = GTRefreshToken
	case "password":
		*gt = GTPassword
	case "client_credentials":
		*gt = ClientCredentials
	default:
		return errors.New("invalid grant_type")
	}

	return nil
}

func (gt *GrantType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "refresh_token":
		*gt = GTRefreshToken
	case "password":
		*gt = GTPassword
	case "client_credentials":
		*gt = ClientCredentials
	default:
		return errors.New("UnmarshalJSON: unknown grant type")
	}

	return nil
}

type ConnectData struct {
	GrantType GrantType `form:"grant_type"`

	RefreshToken string `form:"refresh_token"`

	// Needed for password auth
	ClientID string `form:"client_id"`
	Password string `form:"password"`
	Scope    string `form:"scope"`
	Username string `form:"username"`

	DeviceIdentifier string `form:"deviceIdentifier"`
	DeviceName       string `form:"deviceName"`
	DeviceType       string `form:"deviceType"`
	DevicePushToken  string `form:"devicePushToken"`

	// Needed for two-factor auth
	TwoFactorProvider int32
	TwoFactorToken    string
	TwoFactorRemember int32
}

func (cd ConnectData) Validate() error {
	if cd.GrantType == GTRefreshToken {
		if cd.RefreshToken == "" {
			return errors.New("refresh_token is required")
		}

		return nil
	}

	if cd.ClientID == "" {
		return errors.New("client_id is required")
	}

	if cd.Password == "" {
		return errors.New("password is required")
	}

	if cd.Scope == "" {
		return errors.New("scope is required")
	}

	if cd.Username == "" {
		return errors.New("username is required")
	}

	if cd.DeviceIdentifier == "" {
		return errors.New("device_identifier is required")
	}

	if cd.DeviceName == "" {
		return errors.New("device_name is required")
	}

	if cd.DeviceType == "" {
		return errors.New("device_type is required")
	}

	return nil
}
