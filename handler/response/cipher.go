package response

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/togls/gowarden/model"
)

type Cipher struct {
	Attachments         []*Attachment   `json:"Attachments"`
	Card                json.RawMessage `json:"Card"`
	CollectionIds       []string        `json:"CollectionIds"`
	Data                json.RawMessage `json:"Data"`
	DeletedDate         *time.Time      `json:"DeletedDate"`
	Edit                bool            `json:"Edit"`
	Favorite            bool            `json:"Favorite"`
	Fields              json.RawMessage `json:"Fields"`
	FolderId            *string         `json:"FolderId"`
	Id                  string          `json:"Id"`
	Identity            json.RawMessage `json:"Identity"`
	Login               json.RawMessage `json:"Login"`
	Name                string          `json:"Name"`
	Notes               *string         `json:"Notes"`
	Object              string          `json:"Object"`
	OrganizationId      *string         `json:"OrganizationId"`
	OrganizationUseTotp bool            `json:"OrganizationUseTotp"`
	PasswordHistory     json.RawMessage `json:"PasswordHistory"`
	SecureNote          json.RawMessage `json:"SecureNote"`
	Reprompt            int             `json:"Reprompt"`
	RevisionDate        time.Time       `json:"RevisionDate"`
	Type                int             `json:"Type"`
	ViewPassword        bool            `json:"ViewPassword"`
}

func NewCipher(cipher *model.Cipher, options ...CipherOption) (*Cipher, error) {
	item := &Cipher{
		DeletedDate:         cipher.DeletedAt,
		Id:                  cipher.Uuid,
		Name:                cipher.Name,
		Notes:               cipher.Notes,
		Object:              "cipherDetails",
		OrganizationId:      cipher.OrganizationUuid,
		OrganizationUseTotp: true,
		RevisionDate:        cipher.UpdatedAt,
		Type:                int(cipher.Atype),
	}

	if cipher.Fields != nil {
		item.Fields = json.RawMessage(*cipher.Fields)
	}

	if cipher.PasswordHistory != nil {
		item.PasswordHistory = json.RawMessage(*cipher.PasswordHistory)
	}

	if cipher.Reprompt != nil {
		item.Reprompt = *cipher.Reprompt
	}

	type TypeData struct {
		AutofillOnPageLoad   any
		Fields               json.RawMessage `json:"Fields,omitempty"`
		Name                 string          `json:"Name,omitempty"`
		Notes                *string         `json:"Notes,omitempty"`
		Password             string
		PasswordHistory      *json.RawMessage `json:"PasswordHistory,omitempty"`
		PasswordRevisionDate any
		Totp                 any
		Uri                  *string
		Uris                 []struct {
			Match any
			Uri   string
		}
		Username string
	}
	data := TypeData{}
	if err := json.Unmarshal(cipher.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal cipher data: %w", err)
	}

	if cipher.Atype == 1 {
		if data.Uris != nil {
			data.Uri = &data.Uris[0].Uri
		} else {
			data.Uri = nil
		}
	}

	typeDataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal cipher type data: %w", err)
	}

	switch cipher.Atype {
	case 1:
		item.Login = typeDataBytes
	case 2:
		item.SecureNote = typeDataBytes
	case 3:
		item.Card = typeDataBytes
	case 4:
		item.Identity = typeDataBytes
	default:
		return nil, fmt.Errorf("unknown cipher data type: %d", cipher.Atype)
	}

	if cipher.Fields != nil {
		data.Fields = *cipher.Fields
	}
	data.Name = cipher.Name
	data.Notes = cipher.Notes
	data.PasswordHistory = (*json.RawMessage)(cipher.PasswordHistory)

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal cipher data: %w", err)
	}
	item.Data = dataBytes

	for _, option := range options {
		option.apply(item)
	}

	return item, nil
}

type CipherOption interface {
	apply(item *Cipher)
}

type CipherOptions func(*Cipher)

func (co CipherOptions) apply(item *Cipher) {
	co(item)
}

func CipherWithAttachments(attachments []*model.Attachment, host string) CipherOptions {
	return func(item *Cipher) {
		item.Attachments = NewAttachments(attachments, host)
	}
}

func CipherWithAccess(ro, hp bool) CipherOptions {
	return func(item *Cipher) {
		item.Edit = !ro
		item.ViewPassword = !hp
	}
}

func CipherWithCollectionIds(ids []string) CipherOptions {
	return func(item *Cipher) {
		item.CollectionIds = ids
	}
}

func CipherWithFolderId(folderId string) CipherOptions {
	return func(item *Cipher) {
		item.FolderId = &folderId
	}
}

func CipherWithFavorite(fav bool) CipherOptions {
	return func(item *Cipher) {
		item.Favorite = fav
	}
}
