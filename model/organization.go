package model

import (
	"encoding/json"
)

type Organization struct {
	Uuid         string
	Name         string
	BillingEmail string
	PrivateKey   *string
	PublicKey    *string
}

func (o Organization) MarshalJSON() ([]byte, error) {
	keys := false
	if o.PrivateKey != nil && o.PublicKey != nil {
		keys = true
	}

	data := &struct {
		ID                      string `json:"Id"`
		Name                    string `json:"Name"`
		HasPublicAndPrivateKeys bool   `json:"HasPublicAndPrivateKeys"`
		BillingEmail            string `json:"BillingEmail"`

		Identifier     any  `json:"Identifier"`
		Seats          int  `json:"Seats"`
		MaxCollections int  `json:"MaxCollections"`
		MaxStorageGb   int  `json:"MaxStorageGb"`
		Use2fa         bool `json:"Use2fa"`
		UseDirectory   bool `json:"UseDirectory"`
		UseEvents      bool `json:"UseEvents"`
		UseGroups      bool `json:"UseGroups"`
		UseTotp        bool `json:"UseTotp"`
		UsePolicies    bool `json:"UsePolicies"`
		UseSso         bool `json:"UseSso"`
		SelfHost       bool `json:"SelfHost"`
		UseApi         bool `json:"UseApi"`

		BusinessName      any `json:"BusinessName"`
		BusinessAddress1  any `json:"BusinessAddress1"`
		BusinessAddress2  any `json:"BusinessAddress2"`
		BusinessAddress3  any `json:"BusinessAddress3"`
		BusinessCountry   any `json:"BusinessCountry"`
		BusinessTaxNumber any `json:"BusinessTaxNumber"`

		Plan            string `json:"Plan"`
		PlanType        int    `json:"PlanType"`
		UsersGetPremium bool   `json:"UsersGetPremium"`
		Object          string `json:"Object"`
	}{
		ID:                      o.Uuid,
		Name:                    o.Name,
		HasPublicAndPrivateKeys: keys,
		BillingEmail:            o.BillingEmail,

		Identifier:     nil,
		Seats:          10,
		MaxCollections: 10,
		MaxStorageGb:   10,
		Use2fa:         true,
		UseDirectory:   false,
		UseEvents:      false,
		UseGroups:      false,
		UseTotp:        true,
		UsePolicies:    true,
		UseSso:         false,
		SelfHost:       true,
		UseApi:         false,

		BusinessName:      nil,
		BusinessAddress1:  nil,
		BusinessAddress2:  nil,
		BusinessAddress3:  nil,
		BusinessCountry:   nil,
		BusinessTaxNumber: nil,

		Plan:            "TeamsAnnually",
		PlanType:        5,
		UsersGetPremium: true,
		Object:          "organization",
	}

	return json.Marshal(data)
}
