package model

import (
	"encoding/json"
	"time"
)

type User struct {
	Uuid      string
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time

	VerifiedAt       *time.Time
	LastVerifyingAt  *time.Time
	LoginVerifyCount int

	Email         string
	Name          string
	EmailNew      *string
	EmailNewToken *string

	PasswordHash       []byte
	Salt               []byte
	PasswordIterations int
	PasswordHint       *string

	Akey       *string
	PrivateKey *string
	PublicKey  *string

	TotpSecret  *string
	TotpRecover *string

	SecurityStamp  string
	StampException *string

	EquivalentDomains string
	ExcludedGlobals   string

	ClientKdfType int
	ClientKdfIter int

	ApiKey *string
}

type UserStatus int

const (
	USEnabled UserStatus = iota
	USInvited
	USDisabled
)

func (u *User) ToProfile(uos []*UserOrganization, twofactor, mail bool) (json.RawMessage, error) {
	status := USEnabled
	if string(u.PasswordHash) == "" {
		status = USInvited
	}

	data := struct {
		Culture               string              `json:"Culture"`
		Email                 string              `json:"Email"`
		EmailVerified         bool                `json:"EmailVerified"`
		ForcePasswordReset    bool                `json:"ForcePasswordReset"`
		ID                    string              `json:"Id"`
		Key                   *string             `json:"Key"`
		MasterPasswordHint    *string             `json:"MasterPasswordHint"`
		Name                  string              `json:"Name"`
		Object                string              `json:"Object"`
		Organizations         []*UserOrganization `json:"Organizations"`
		Premium               bool                `json:"Premium"`
		PrivateKey            *string             `json:"PrivateKey"`
		ProviderOrganizations json.RawMessage     `json:"ProviderOrganizations"`
		Providers             json.RawMessage     `json:"Providers"`
		TwoFactor             bool                `json:"TwoFactorEnabled"`
		SecurityStamp         string              `json:"SecurityStamp"`
		Status                int                 `json:"_Status"`
	}{
		Status:                int(status),
		ID:                    u.Uuid,
		Name:                  u.Name,
		Email:                 u.Email,
		EmailVerified:         (!mail || u.VerifiedAt != nil),
		Premium:               true,
		MasterPasswordHint:    u.PasswordHint,
		Culture:               "en-US",
		TwoFactor:             twofactor,
		Key:                   u.Akey,
		PrivateKey:            u.PrivateKey,
		SecurityStamp:         u.SecurityStamp,
		Organizations:         uos,
		Providers:             json.RawMessage(`[]`),
		ProviderOrganizations: json.RawMessage(`[]`),
		ForcePasswordReset:    false,
		Object:                "profile",
	}

	return json.Marshal(&data)
}

type UpdateUser struct {
	Uuid string // must be set

	Name           *string
	PasswordHash   []byte
	PasswordHint   *string
	Akey           *string
	PublicKey      *string
	PrivateKey     *string
	ClientKdfType  *int
	ClientKdfIter  *int
	SecurityStamp  *string
	StampException *string
	Email          *string
	EmailNew       *string
	EmailNewToken  *string
	ApiKey         *string

	VerifiedAt       *time.Time
	LastVerifyingAt  *time.Time
	LoginVerifyCount *int

	UpdatedAt *time.Time
}

const (
	ClientKdfTypeDefault int = 0 // PBKDF2: 0
	ClientKdfIterDefault int = 100_000
)

type UserStampException struct {
	Routes        []string  `json:"routes"`
	SecurityStamp string    `json:"security_stamp"`
	Expire        time.Time `json:"expire"`
}

type Invitation struct {
	Email string
}
