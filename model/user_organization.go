package model

type UserOrganization struct {
	Uuid     string
	UserUuid string
	OrgUuid  string

	AccessAll bool
	AKey      *string

	Status UOStatus
	Atype  UOType

	Name       string
	PrivateKey *string
	PublicKey  *string
}

type UOFilter struct {
	UserUuid *string
	OrgUuid  *string
	Status   *UOStatus
	Atype    *UOType
}

// Invited = 0
// Accepted = 1
// Confirmed = 2
type UOStatus int

const (
	UOStatusInvited UOStatus = iota
	UOStatusAccepted
	UOStatusConfirmed
)

// Owner = 0
// Admin = 1
// User = 2
// Manager = 3
type UOType int

const (
	UOTypeOwner UOType = iota
	UOTypeAdmin
	UOTypeUser
	UOTypeManager
)

type UODetail struct {
	ID          string          `json:"Id"`
	UserID      string          `json:"UserId"`
	Status      int             `json:"Status"`
	Type        int             `json:"Type"`
	AccessAll   bool            `json:"AccessAll"`
	Collections []*UOCollection `json:"Collections"`
	Object      string          `json:"Object"`
}

type UOCollection struct {
	ID            string `json:"Id"`
	ReadOnly      bool   `json:"ReadOnly"`
	HidePasswords bool   `json:"HidePasswords"`
}
