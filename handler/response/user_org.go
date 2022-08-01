package response

import "github.com/togls/gowarden/model"

type UserOrganization struct {
	Enabled                 bool    `json:"Enabled"`
	HasPublicAndPrivateKeys bool    `json:"HasPublicAndPrivateKeys"`
	Id                      string  `json:"Id"`
	Identifier              any     `json:"Identifier"`
	Key                     *string `json:"Key"`
	MaxCollections          int     `json:"MaxCollections"`
	MaxStorageGb            int     `json:"MaxStorageGb"`
	Name                    string  `json:"Name"`
	Object                  string  `json:"Object"`
	ProviderId              any     `json:"ProviderId"`
	ProviderName            any     `json:"ProviderName"`
	ResetPasswordEnrolled   bool    `json:"ResetPasswordEnrolled"`
	Seats                   int     `json:"Seats"`
	SelfHost                bool    `json:"SelfHost"`
	SsoBound                bool    `json:"SsoBound"`
	Status                  int     `json:"Status"`
	Type                    int     `json:"Type"`
	Use2fa                  bool    `json:"Use2fa"`
	UsersGetPremium         bool    `json:"UsersGetPremium"`
	UseDirectory            bool    `json:"UseDirectory"`
	UseEvents               bool    `json:"UseEvents"`
	UseGroups               bool    `json:"UseGroups"`
	UseTotp                 bool    `json:"UseTotp"`
	UsePolicies             bool    `json:"UsePolicies"`
	UseApi                  bool    `json:"UseApi"`
	UseSso                  bool    `json:"UseSso"`
	UseBusinessPortal       bool    `json:"UseBusinessPortal"`
}

func NewUserOrganization(userOrg *model.UserOrganization) *UserOrganization {
	return &UserOrganization{
		Id:              userOrg.OrgUuid,
		Identifier:      nil,
		Name:            userOrg.Name,
		Seats:           10,
		MaxCollections:  10,
		UsersGetPremium: true,

		Use2fa:                  true,
		UseDirectory:            false,
		UseEvents:               false,
		UseGroups:               false,
		UseTotp:                 true,
		UsePolicies:             true,
		UseApi:                  false,
		SelfHost:                true,
		HasPublicAndPrivateKeys: (userOrg.PrivateKey != nil && userOrg.PublicKey != nil),
		ResetPasswordEnrolled:   false,
		SsoBound:                false,
		UseSso:                  false,
		UseBusinessPortal:       false,
		ProviderId:              nil,
		ProviderName:            nil,

		MaxStorageGb: 10,

		Key:     userOrg.AKey,
		Status:  int(userOrg.Status),
		Type:    int(userOrg.Atype),
		Enabled: true,

		Object: "profileOrganization",
	}
}

func NewUserOrganizations(userOrgs []*model.UserOrganization) []*UserOrganization {
	var result []*UserOrganization
	for _, userOrg := range userOrgs {
		result = append(result, NewUserOrganization(userOrg))
	}
	return result
}
