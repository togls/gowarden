package handler

import (
	"encoding/json"
	"time"

	"github.com/togls/gowarden/model"
	"github.com/togls/gowarden/pkg/crypto"
)

type CipherData struct {
	ID             *string `json:"Id"`
	FolderId       *string `json:"FolderId"`
	OrganizationId *string `json:"OrganizationId"`

	Type  int     `json:"Type"`
	Name  string  `json:"Name"`
	Notes *string `json:"Notes"`

	Favorite *bool `json:"Favorite"`
	Reprompt *int  `json:"Reprompt"`

	Attachments  *json.RawMessage `json:"Attachments"`
	Attachments2 *json.RawMessage `json:"Attachments2"`

	LastKnownRevisionDate time.Time `json:"LastKnownRevisionDate"`

	Fields          *json.RawMessage `json:"Fields"` // Raw Json string
	PasswordHistory *json.RawMessage `json:"PasswordHistory"`

	Login      json.RawMessage `json:"Login"`
	SecureNote json.RawMessage `json:"SecureNote"`
	Card       json.RawMessage `json:"Card"`
	Identity   json.RawMessage `json:"Identity"`
}

func (cd CipherData) toCipher() (*model.Cipher, error) {
	c := &model.Cipher{
		Atype:            model.CipherType(cd.Type),
		Name:             cd.Name,
		UpdatedAt:        cd.LastKnownRevisionDate,
		OrganizationUuid: cd.OrganizationId,
		Notes:            cd.Notes,
		Reprompt:         cd.Reprompt,
	}

	if cd.ID != nil {
		c.Uuid = *cd.ID
	} else {
		uuid, err := crypto.GenerateUuid()
		if err != nil {
			return nil, err
		}
		c.Uuid = uuid
	}

	if cd.Fields != nil {
		c.Fields = (*[]byte)(cd.Fields)
	}

	if cd.PasswordHistory != nil {
		c.PasswordHistory = (*[]byte)(cd.PasswordHistory)
	}

	// TODO: handle attachments

	// TODO: remove field 'response' in the raw json
	switch cd.Type {
	case 1:
		c.Data = cd.Login
	case 2:
		c.Data = cd.SecureNote
	case 3:
		c.Data = cd.Card
	case 4:
		c.Data = cd.Identity
	}

	return c, nil
}
