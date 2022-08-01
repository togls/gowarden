package model

type OrgPolicy struct {
	Uuid    string
	OrgUuid string
	Atype   OrgPolicyType
	Enabled bool
	Data    []byte
}

type OrgPolicyType int

const (
	OPTypeTwoFactor OrgPolicyType = 0 + iota
	OPTypeMasterPassword
	OPTypePasswordGenerator
	OPTypeSingleOrg
	OPTypeRequireSso
	OPTypePersonalOwnership
	OPTypeDisableSend
	OPTypeSendOptions
)
