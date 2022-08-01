package model

import (
	"encoding/json"
	"time"
)

type Folder struct {
	Uuid     string `json:"Id"`
	UserUuid string `json:"-"`
	Name     string `json:"Name"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"RevisionDate"`
}

func (f *Folder) MarshalJSON() ([]byte, error) {
	type Alias Folder

	data := struct {
		*Alias
		Object string `json:"Object"`
	}{
		Alias:  (*Alias)(f),
		Object: "folder",
	}

	return json.Marshal(&data)
}

type FolderCipher struct {
	FolderUuid string
	CipherUuid string
}

type NewFolder struct {
	Name string // hashed
}
