package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/togls/gowarden/auth"
	"github.com/togls/gowarden/model"
	"github.com/togls/gowarden/pkg/crypto"
	"github.com/togls/gowarden/store"
)

type AccountHandler struct {
	users   store.User
	devices store.Device
	uos     store.UserOrganization
	sends   store.Send
	eas     store.EmergencyAccess
	ciphers store.Cipher
	fas     store.Favorite
	fs      store.Folder
	tfs     store.TwoFactor
	tfis    store.TwoFactorIncomplete
	is      store.Invitation

	auth *auth.Core
	dec  auth.JWTDecoder
}

func NewAccountHandler(
	users store.User,
	devices store.Device,
	uos store.UserOrganization,
	sends store.Send,
	eas store.EmergencyAccess,
	ciphers store.Cipher,
	fas store.Favorite,
	fs store.Folder,
	tfs store.TwoFactor,
	tfis store.TwoFactorIncomplete,
	is store.Invitation,
	auth *auth.Core,
	dec auth.JWTDecoder,
) *AccountHandler {
	return &AccountHandler{
		users:   users,
		devices: devices,
		uos:     uos,
		sends:   sends,
		eas:     eas,
		ciphers: ciphers,
		fas:     fas,
		fs:      fs,
		tfs:     tfs,
		tfis:    tfis,
		is:      is,
		auth:    auth,
		dec:     dec,
	}
}

func (ah AccountHandler) Routes(e *echo.Echo) {
	e.POST("/api/accounts/register", ah.Register)
	e.POST("/api/accounts/verify-email-token", ah.PostVerifyEmailToken)
	e.POST("/delete-recover", ah.PostDeleteRecover)
	e.POST("/delete-recover-token", ah.PostDeleteRecoverToken)

	e.GET("/api/users/:uuid/public-key", ah.GetPublicKeys, ah.auth.RequireAuth)

	e.POST("/api/accounts/prelogin", ah.Prelogin)
	e.POST("/api/accounts/password-hint", ah.PasswordHint)

	account := e.Group("/api/accounts", ah.auth.RequireAuth)

	account.GET("/profile", ah.Profile)
	account.PUT("/profile", ah.PutProfile)
	account.POST("/profile", ah.PostProfile)

	account.POST("/keys", ah.PostKeys)
	account.POST("/password", ah.PostPassword)
	account.POST("/kdf", ah.PostKdf)
	account.POST("/key", ah.PostRotateKey)

	account.POST("/security-stamp", ah.PostSecurityStamp)
	account.POST("/email-token", ah.PostEmailToken)
	account.POST("/email", ah.PostEmail)
	account.POST("/verify-email", ah.PostVerifyEmail)

	account.DELETE("", ah.DeleteAccount)
	account.POST("/delete", ah.PostDeleteAccount)
	account.GET("/revision-date", ah.RevisionDate)
	account.POST("/verify-password", ah.VerifyPassword)
	account.POST("/api-key", ah.ApiKey)
	account.POST("/rotate-api-key", ah.RotateApiKey)

}

type PreloginData struct {
	Email string `json:"email"`
}

func (ah *AccountHandler) Prelogin(c echo.Context) error {
	var pd PreloginData
	if err := c.Bind(&pd); err != nil {
		return err
	}

	typ := model.ClientKdfTypeDefault
	iter := model.ClientKdfIterDefault

	u, err := ah.users.FindByEmail(pd.Email)
	if err == nil && u != nil {
		typ = u.ClientKdfType
		iter = u.ClientKdfIter
	}

	resp := struct {
		Kdf           int `json:"Kdf"`
		KdfIterations int `json:"KdfIterations"`
	}{
		Kdf:           typ,
		KdfIterations: iter,
	}

	return c.JSON(http.StatusOK, resp)
}

type RegisterData struct {
	Email              string    `json:"email"`
	Kdf                *int      `json:"kdf"`
	KdfIterations      *int      `json:"kdfIterations"`
	Key                *string   `json:"key"`
	Keys               *KeysData `json:"keys"`
	MasterPasswordHash string    `json:"masterPasswordHash"`
	MasterPasswordHint *string   `json:"masterPasswordHint"`
	Name               *string   `json:"name"`
	Token              string    `json:"token"`
	OrganizationUserId string    `json:"organizationUserId"`
}

func (rd *RegisterData) toUser() (*model.User, error) {
	now := time.Now()

	u := &model.User{
		Enabled:            true,
		CreatedAt:          now,
		UpdatedAt:          now,
		Name:               rd.Email,
		Email:              rd.Email,
		ClientKdfType:      model.ClientKdfTypeDefault,
		ClientKdfIter:      model.ClientKdfIterDefault,
		PasswordIterations: 100_000,
	}

	userUuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	u.Uuid = userUuid.String()

	stUuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	u.SecurityStamp = stUuid.String()

	if rd.Kdf != nil {
		u.ClientKdfType = *rd.Kdf
	}

	if rd.KdfIterations != nil {
		u.ClientKdfIter = *rd.KdfIterations
	}

	if rd.Name != nil {
		u.Name = *rd.Name
	} else {
		return nil, errors.New("name is required")
	}

	u.Akey = rd.Key

	salt, err := crypto.GenerateBytes(64)
	if err != nil {
		return nil, err
	}

	pwHash := crypto.GeneratePassword(
		rd.MasterPasswordHash, salt,
		u.PasswordIterations)

	u.PasswordHash = pwHash
	u.Salt = salt
	u.PasswordHint = rd.MasterPasswordHint

	if rd.Keys != nil {
		u.PublicKey = &rd.Keys.PublicKey
		u.PrivateKey = &rd.Keys.EncryptedPrivateKey
	}

	return u, nil
}

type KeysData struct {
	EncryptedPrivateKey string `json:"encryptedPrivateKey"`
	PublicKey           string `json:"publicKey"`
}

func (ah *AccountHandler) Register(c echo.Context) error {
	data := new(RegisterData)
	if err := c.Bind(data); err != nil {
		return err
	}

	exited, err := ah.users.FindByEmail(data.Email)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return err
	}

	// TODO is_signup_allowed

	// check if user already exists
	if exited != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "User already exists")
	}

	newUser, err := data.toUser()
	if err != nil {
		return err
	}

	if err := ah.users.Create(newUser); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (ah *AccountHandler) Profile(c echo.Context) error {

	user := auth.GetUser(c)

	return c.JSON(http.StatusOK, user)
}

type ProfileData struct {
	MasterPasswordHint *string `json:"masterPasswordHint"`
	Name               string  `json:"name"`
}

func (ah *AccountHandler) PostProfile(c echo.Context) error {
	data := new(ProfileData)
	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)
	updateUser := &model.UpdateUser{
		Uuid: user.Uuid,
		Name: &data.Name,
	}

	if data.MasterPasswordHint != nil {
		updateUser.PasswordHint = data.MasterPasswordHint
	}

	if err := ah.users.Update(updateUser); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (ah *AccountHandler) PutProfile(c echo.Context) error {
	return ah.PostProfile(c)
}

func (ah *AccountHandler) GetPublicKeys(c echo.Context) error {
	uuid := c.Param("uuid")

	user, err := ah.users.FindByUuid(uuid)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "User doesn't exist")
	}

	resp := &struct {
		UserID    string  `json:"UserId"`
		PublicKey *string `json:"PublicKey"`
		Object    string  `json:"userKey"`
	}{
		UserID:    user.Uuid,
		PublicKey: user.PublicKey,
		Object:    "userKey",
	}

	return c.JSON(http.StatusOK, resp)
}

func (ah *AccountHandler) PostKeys(c echo.Context) error {
	data := new(KeysData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)

	uu := &model.UpdateUser{
		Uuid: user.Uuid,

		PublicKey:  &data.PublicKey,
		PrivateKey: &data.EncryptedPrivateKey,
	}

	if err := ah.users.Update(uu); err != nil {
		return err
	}

	resp := &struct {
		PublicKey  string `json:"PublicKey"`
		PrivateKey string `json:"PrivateKey"`
		Object     string `json:"keys"`
	}{
		PublicKey:  data.PublicKey,
		PrivateKey: data.EncryptedPrivateKey,
		Object:     "keys",
	}

	return c.JSON(http.StatusOK, resp)
}

type UpdatePWData struct {
	MasterPasswordHash    string `json:"masterPasswordHash"`
	NewMasterPasswordHash string `json:"newMasterPasswordHash"`
	Key                   string `json:"key"`
}

func (ah *AccountHandler) PostPassword(c echo.Context) error {
	data := new(UpdatePWData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)

	ok := crypto.VerifyPassword(
		data.MasterPasswordHash,
		user.Salt,
		user.PasswordHash,
		user.PasswordIterations,
	)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid password")
	}

	pwHash := crypto.GeneratePassword(
		data.NewMasterPasswordHash,
		user.Salt,
		user.PasswordIterations)

	uu := &model.UpdateUser{
		Uuid:         user.Uuid,
		PasswordHash: pwHash,
		Akey:         &data.Key,
	}

	if err := ah.users.Update(uu); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

type UpdateKdfData struct {
	Kdf           int `json:"kdf"`
	KdfIterations int `json:"kdfIterations"`

	UpdatePWData
}

func (ah *AccountHandler) PostKdf(c echo.Context) error {
	data := new(UpdateKdfData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)

	ok := crypto.VerifyPassword(
		data.MasterPasswordHash,
		user.Salt,
		user.PasswordHash,
		user.PasswordIterations,
	)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid password")
	}

	pwHash := crypto.GeneratePassword(
		data.NewMasterPasswordHash, user.Salt, data.KdfIterations)

	uu := &model.UpdateUser{
		Uuid:          user.Uuid,
		ClientKdfType: &data.Kdf,
		ClientKdfIter: &data.KdfIterations,
		Akey:          &data.Key,
		PasswordHash:  pwHash,
	}

	if err := ah.users.Update(uu); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (ah *AccountHandler) PostRotateKey(c echo.Context) error {
	// Cipher
	// Folder
	panic("TODO: not implemented")
}

type PasswordData struct {
	MasterPasswordHash string `json:"masterPasswordHash"`
}

func (ah *AccountHandler) PostSecurityStamp(c echo.Context) error {
	data := new(PasswordData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)

	ok := crypto.VerifyPassword(
		data.MasterPasswordHash,
		user.Salt,
		user.PasswordHash,
		user.PasswordIterations,
	)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid password")
	}

	if err := ah.devices.DeleteAllByUser(user.Uuid); err != nil {
		return err
	}

	ssr, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	ssUuid := ssr.String()
	updateUser := &model.UpdateUser{
		Uuid:          user.Uuid,
		SecurityStamp: &ssUuid,
	}

	if err := ah.users.Update(updateUser); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

type EmailTokenData struct {
	MasterPasswordHash string `json:"MasterPasswordHash"`
	NewEmail           string `json:"NewEmail"`
}

func (ah *AccountHandler) PostEmailToken(c echo.Context) error {
	data := new(EmailTokenData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)

	ok := crypto.VerifyPassword(
		data.MasterPasswordHash,
		user.Salt,
		user.PasswordHash,
		user.PasswordIterations,
	)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid password")
	}

	// TODO: is_email_domain_allowed

	token, err := crypto.GenerateAlphanumString(6)
	if err != nil {
		return err
	}

	// TODO: mail_enabled

	uu := &model.UpdateUser{
		Uuid: user.Uuid,

		EmailNew:      &data.NewEmail,
		EmailNewToken: &token,
	}
	if err := ah.users.Update(uu); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

type ChangeEmailData struct {
	MasterPasswordHash string `json:"MasterPasswordHash"`
	NewEmail           string `json:"NewEmail"`

	Key                   string `json:"Key"`
	NewMasterPasswordHash string `json:"NewMasterPasswordHash"`
	Token                 string `json:"Token"`
}

func (ah *AccountHandler) PostEmail(c echo.Context) error {
	data := new(ChangeEmailData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)

	ok := crypto.VerifyPassword(
		data.MasterPasswordHash,
		user.Salt,
		user.PasswordHash,
		user.PasswordIterations,
	)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid password")
	}

	exist, err := ah.users.FindByEmail(data.NewEmail)
	if err != nil {
		return err
	}

	if exist != nil {
		return echo.NewHTTPError(http.StatusConflict, "Email already in use")
	}

	if user.EmailNew == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No email change pending")
	}

	if user.EmailNewToken == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Email change mismatch")
	}

	// TODO: mail_enabled
	// check token

	pwHash := crypto.GeneratePassword(
		data.NewMasterPasswordHash,
		user.Salt,
		user.PasswordIterations,
	)

	empty := ""
	uu := &model.UpdateUser{
		Uuid: user.Uuid,

		PasswordHash:  pwHash,
		Email:         &data.NewEmail,
		EmailNew:      &empty,
		EmailNewToken: &empty,
		Akey:          &data.Key,
	}

	if err := ah.users.Update(uu); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (ah *AccountHandler) PostVerifyEmail(c echo.Context) error {

	// TODO: mail_enabled

	return c.NoContent(http.StatusOK)
}

type VerifyEmailTokenData struct {
	UserId string `json:"UserId"`
	Token  string `json:"Token"`
}

func (ah *AccountHandler) PostVerifyEmailToken(c echo.Context) error {
	data := new(VerifyEmailTokenData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user, err := ah.users.FindByUuid(data.UserId)
	if err != nil {
		return err
	}
	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	claims := &jwt.RegisteredClaims{}

	if err := ah.dec.DecodeToken(data.Token, claims); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
	}

	if claims.Issuer != "|verifyemail" || claims.Subject != user.Uuid {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
	}

	now := time.Now()
	zero := 0
	uu := model.UpdateUser{
		Uuid: user.Uuid,

		VerifiedAt:       &now,
		LastVerifyingAt:  &time.Time{},
		LoginVerifyCount: &zero,
	}

	if err := ah.users.Update(&uu); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

type DeleteRecoverData struct {
	Email string `json:"Email"`
}

func (ah *AccountHandler) PostDeleteRecover(c echo.Context) error {
	data := new(DeleteRecoverData)

	if err := c.Bind(data); err != nil {
		return err
	}

	_, err := ah.users.FindByEmail(data.Email)
	if err != nil {
		return err
	}

	// TODO: mail_enabled

	return echo.NewHTTPError(http.StatusBadRequest, "Please contact the administrator to delete your account")
}

type DeleteRecoverTokenData struct {
	UserID string `json:"UserId"`
	Token  string `json:"Token"`
}

func (ah *AccountHandler) PostDeleteRecoverToken(c echo.Context) error {
	data := new(DeleteRecoverTokenData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user, err := ah.users.FindByUuid(data.UserID)
	if err != nil {
		return err
	}

	claims := &jwt.RegisteredClaims{}
	if err := ah.dec.DecodeToken(data.Token, claims); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
	}

	if claims.Issuer != "|delete" || claims.Subject != user.Uuid {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
	}

	if err := ah.deleteUser(user); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (ah *AccountHandler) DeleteAccount(c echo.Context) error {
	data := new(PasswordData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)

	ok := crypto.VerifyPassword(
		data.MasterPasswordHash,
		user.Salt,
		user.PasswordHash,
		user.PasswordIterations,
	)

	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid password")
	}

	if err := ah.deleteUser(user); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (ah *AccountHandler) deleteUser(user *model.User) error {

	c := model.UOStatusConfirmed
	filter := &model.UOFilter{
		UserUuid: &user.Uuid,
		Status:   &c,
	}

	uos, err := ah.uos.Find(filter)
	if err != nil {
		return err
	}

	for _, uo := range uos {
		if uo.Atype != model.UOTypeOwner {
			continue
		}

		owner := model.UOTypeOwner
		f := &model.UOFilter{
			OrgUuid: &uo.OrgUuid,
			Atype:   &owner,
		}
		os, err := ah.uos.Find(f)
		if err != nil {
			return err
		}

		if len(os) <= 1 {
			return echo.NewHTTPError(http.StatusBadRequest, "Cann't delete last owner")
		}
	}

	if err := ah.sends.DeleteAllByUser(user.Uuid); err != nil {
		return err
	}

	if err := ah.eas.DeleteAllByUser(user.Uuid); err != nil {
		return err
	}

	if err := ah.uos.DeleteAllByUser(user.Uuid); err != nil {
		return err
	}

	if err := ah.ciphers.DeleteByUser(user.Uuid); err != nil {
		return err
	}

	if err := ah.fas.DeleteAllByUser(user.Uuid); err != nil {
		return err
	}

	if err := ah.fs.DeleteAllByUser(user.Uuid); err != nil {
		return err
	}

	if err := ah.devices.DeleteAllByUser(user.Uuid); err != nil {
		return err
	}

	if err := ah.tfs.DeleteAllByUser(user.Uuid); err != nil {
		return err
	}

	if err := ah.tfis.DeleteAllByUser(user.Uuid); err != nil {
		return err
	}

	if err := ah.is.Delete(user.Email); err != nil {
		return err
	}

	return nil
}

func (ah *AccountHandler) PostDeleteAccount(c echo.Context) error {
	return ah.DeleteAccount(c)
}

func (ah *AccountHandler) RevisionDate(c echo.Context) error {
	user := auth.GetUser(c)

	return c.String(http.StatusOK, fmt.Sprintf("%d", user.UpdatedAt.UnixMilli()))
}

type PasswordHintData struct {
	Email string `json:"Email"`
}

func (ah *AccountHandler) PasswordHint(c echo.Context) error {
	data := new(PasswordHintData)

	if err := c.Bind(data); err != nil {
		return err
	}

	// TODO: mail_enabled
	return c.String(http.StatusBadRequest, "This server is not configured to provide password hints.")
}

func (ah *AccountHandler) VerifyPassword(c echo.Context) error {
	data := new(PasswordData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)

	ok := crypto.VerifyPassword(
		data.MasterPasswordHash,
		user.Salt,
		user.PasswordHash,
		user.PasswordIterations,
	)

	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid password")
	}

	return c.NoContent(http.StatusOK)
}

func (ah *AccountHandler) ApiKey(c echo.Context) error {
	return ah.apiKey(c, false)
}

func (ah *AccountHandler) RotateApiKey(c echo.Context) error {
	return ah.apiKey(c, true)
}

func (ah *AccountHandler) apiKey(c echo.Context, rotate bool) error {
	data := new(PasswordData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)

	ok := crypto.VerifyPassword(
		data.MasterPasswordHash,
		user.Salt,
		user.PasswordHash,
		user.PasswordIterations,
	)

	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid password")
	}

	if rotate || user.ApiKey == nil {
		ak, err := crypto.GenerateApiKey()
		if err != nil {
			return err
		}

		user.ApiKey = &ak
		uu := &model.UpdateUser{
			Uuid: user.Uuid,

			ApiKey: user.ApiKey,
		}
		if err := ah.users.Update(uu); err != nil {
			return err
		}
	}

	resp := struct {
		ApiKey *string `json:"ApiKey"`
		Object string  `json:"Object"`
	}{
		ApiKey: user.ApiKey,
		Object: "apiKey",
	}

	return c.JSON(http.StatusOK, resp)
}
