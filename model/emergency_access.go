package model

import (
	"database/sql"
	"time"
)

type EmergencyAccess struct {
	Uuid         string `json:"Id"`
	Status       int    `json:"Status"`
	Atype        int    `json:"Type"`
	WaitTimeDays int    `json:"WaitTimeDays"`

	GrantorUuid *string `json:"-"`
	GranteeUuid *string `json:"-"`
	Email       *string `json:"-"`
	Name        string  `json:"-"`

	KeyEncrypted        *string      `json:"-"`
	RecoveryInitiatedAt sql.NullTime `json:"-"`
	LastNotificationAt  sql.NullTime `json:"-"`
	UpdatedAt           time.Time    `json:"-"`
	CreatedAt           time.Time    `json:"-"`
}

type EAFilter struct {
	GrantorUuid *string
	GranteeUuid *string
}
