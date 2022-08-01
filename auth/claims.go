package auth

import (
	"github.com/golang-jwt/jwt/v4"
)

type LoginJwtClaims struct {
	// NotBefore int64  `json:"nbf"` // Not before
	// ExpiresAt int64  `json:"exp"` // Expiration time
	// Issuer    string `json:"iss"` // Issuer
	// Subject   string `json:"sub"` // Subject
	*jwt.RegisteredClaims

	Device        string `json:"device"` // device uuid
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`

	Orgowner   []string `json:"orgowner"`
	Orgadmin   []string `json:"orgadmin"`
	Orguser    []string `json:"orguser"`
	Orgmanager []string `json:"orgmanager"`

	Premium bool     `json:"premium"`
	Sstamp  string   `json:"sstamp"` // user security_stamp
	Scope   []string `json:"scope"`  // [ "api", "offline_access" ]
	Amr     []string `json:"amr"`    // [ "Application" ]
}
