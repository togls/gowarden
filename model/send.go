package model

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

type Send struct {
	Uuid             string
	UserUuid         *string
	OrganizationUuid *string

	Name  string
	Notes *string
	Atype SendType
	Data  []byte
	Akey  string

	PasswordHash *[]byte
	PasswordSalt *[]byte
	PasswordIter *int

	MaxAccessCount *int
	AccessCount    int

	CreationDate   time.Time
	RevisionDate   time.Time
	ExpirationDate *time.Time
	DeletionDate   time.Time

	Disabled  bool
	HideEmail *bool
}

func (item Send) MarshalJSON() ([]byte, error) {
	type Alias Send
	aux := &struct {
		*Alias
		AccessId string          `json:"AccessId"`
		Text     json.RawMessage `json:"Text"`
		File     json.RawMessage `json:"File"`
		Object   string          `json:"Object"`
	}{
		Alias:  (*Alias)(&item),
		Object: "send",
	}

	id, err := base64.RawURLEncoding.DecodeString(item.Uuid)
	if err != nil {
		return nil, fmt.Errorf("invalid send id: %s, MarshalJSON err: %w", item.Uuid, err)
	}

	aux.AccessId = string(id)

	data, err := json.Marshal(item.Data)
	if err != nil {
		return nil, fmt.Errorf("Send MarshalJSON err: %w", err)
	}

	if item.Atype == SendTypeText {
		aux.Text = data
	} else if item.Atype == SendTypeFile {
		aux.File = data
	}

	return json.Marshal(aux)
}

// Text = 0
// File = 1
type SendType int

const (
	SendTypeText SendType = iota
	SendTypeFile
)

type SendFilter struct {
	UserUuid *string
}
