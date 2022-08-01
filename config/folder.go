package config

import (
	"errors"
	"fmt"
	"os"
)

type Folders struct {
	Data           string
	DatabaseURL    string `json:"database_url"`
	IconCache      string
	Attachments    string
	Sends          string
	Tmp            string
	Templates      string
	RsaKeyFilename string
	Web            string `json:"web"`
}

func (f *Folders) check() error {
	if f.Data == "" {
		f.Data = "/app/data"
	}

	if f.DatabaseURL == "" {
		f.DatabaseURL = f.Data + "/db.sqlite3"
	}

	f.IconCache = f.Data + "/icon_cache"
	f.Attachments = f.Data + "/attachments"
	f.Sends = f.Data + "/sends"
	f.Tmp = f.Data + "/tmp"
	f.Templates = f.Data + "/templates"
	f.RsaKeyFilename = f.Data + "/rsa_key"

	if f.Web == "" {
		f.Web = "web-vault"
	}

	_, err := os.Stat(f.Data)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(f.Data, os.ModeDir); err != nil {
			return fmt.Errorf("failed to create data folder: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to stat data folder: %w", err)
	}

	return nil
}

func (f *Folders) PrivateKeyPath() string {
	return f.RsaKeyFilename + ".pem"
}

func (f *Folders) PublicKeyPath() string {
	return f.RsaKeyFilename + ".pub.pem"
}
