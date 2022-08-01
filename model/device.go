package model

import (
	"time"
)

type Device struct {
	Uuid      string
	CreatedAt time.Time
	UpdatedAt time.Time

	UserUuid string

	Name      string
	Atype     int
	PushToken *string

	RefreshToken string

	TwofactorRemember *string
}
