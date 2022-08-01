package model

type TwoFactor struct {
	Uuid     string
	UserUuid string
	Atype    int
	Enabled  bool
	Data     string
	LastUsed int
}

// TwoFactorType
// Authenticator = 0
// Email = 1
// Duo = 2
// YubiKey = 3
// U2f = 4
// Remember = 5
// OrganizationDuo = 6
// Webauthn = 7
//
// // These are implementation details
// U2fRegisterChallenge = 1000
// U2fLoginChallenge = 1001
// EmailVerificationChallenge = 1002
// WebauthnRegisterChallenge = 1003
// WebauthnLoginChallenge = 1004
