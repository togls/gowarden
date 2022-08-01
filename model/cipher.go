package model

import (
	"time"
)

type Cipher struct {
	Uuid             string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	UserUuid         *string
	OrganizationUuid *string
	Atype            CipherType
	Name             string
	Notes            *string
	Fields           *[]byte
	Data             []byte
	PasswordHistory  *[]byte
	DeletedAt        *time.Time
	Reprompt         *int // None = 0, Password = 1 (not used)

	// Associations
	Attachments   []*Attachment
	ReadOnly      bool
	HidePassword  bool
	CollectionIDs []string
	FolderId      *string
	Favorite      bool
}

type CipherType int

const (
	CTypeLogin CipherType = 1 + iota
	CTypeSecureNote
	CTypeCard
	CTypeIdentity
)

type UpdateCipher struct {
	Uuid string

	DeletedAt *time.Time
}
