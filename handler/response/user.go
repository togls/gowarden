package response

import (
	"encoding/json"

	"github.com/togls/gowarden/model"
)

type Profile struct {
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
}

func NewProfile(u *model.User, uos []*model.UserOrganization, twofactor, mail bool) *Profile {
	status := model.USEnabled
	if string(u.PasswordHash) == "" {
		status = model.USInvited
	}

	return &Profile{
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
		Organizations:         NewUserOrganizations(uos),
		Providers:             json.RawMessage(`[]`),
		ProviderOrganizations: json.RawMessage(`[]`),
		ForcePasswordReset:    false,
		Object:                "profile",
	}
}
