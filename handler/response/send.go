package response

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/togls/gowarden/model"
)

type Send struct {
	AccessCount    int             `json:"AccessCount"`
	AccessId       string          `json:"AccessId"`
	DeletionDate   time.Time       `json:"DeletionDate"`
	Disabled       bool            `json:"Disabled"`
	ExpirationDate *time.Time      `json:"ExpirationDate"`
	File           json.RawMessage `json:"File"`
	HideEmail      *bool           `json:"HideEmail"`
	Id             string          `json:"Id"`
	Key            string          `json:"Key"`
	MaxAccessCount *int            `json:"MaxAccessCount"`
	Name           string          `json:"Name"`
	Notes          *string         `json:"Notes"`
	Object         string          `json:"Object"`
	Password       *string         `json:"Password"`
	RevisionDate   time.Time       `json:"RevisionDate"`
	Text           json.RawMessage `json:"Text"`
	Type           int             `json:"Type"`
}

func NewSend(src *model.Send) (*Send, error) {
	aux := Send{
		Object: "send",
	}

	id, err := base64.RawURLEncoding.DecodeString(src.Uuid)
	if err != nil {
		return nil, fmt.Errorf("NewSend invalid send id: %s, MarshalJSON err: %w", src.Uuid, err)
	}

	aux.AccessId = string(id)

	data, err := json.Marshal(src.Data)
	if err != nil {
		return nil, fmt.Errorf("NewSend MarshalJSON err: %w", err)
	}

	switch src.Atype {
	case model.SendTypeText:
		aux.Text = data
	case model.SendTypeFile:
		aux.File = data
	default:
		return nil, fmt.Errorf("NewSend invalid send type: %d", src.Atype)
	}

	return &aux, nil
}

func NewSends(src []*model.Send) ([]*Send, error) {
	var result []*Send
	for _, item := range src {
		aux, err := NewSend(item)
		if err != nil {
			return nil, err
		}
		result = append(result, aux)
	}
	return result, nil
}
