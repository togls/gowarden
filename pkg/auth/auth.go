package auth

import (
	"crypto/rsa"
	"errors"

	"github.com/golang-jwt/jwt/v4"
)

type LoginJwtClaims struct {
	// NotBefore int64  `json:"nbf"` // Not before
	// ExpiresAt int64  `json:"exp"` // Expiration time
	// Issuer    string `json:"iss"` // Issuer
	// Subject   string `json:"sub"` // Subject
	*jwt.RegisteredClaims

	Premium       bool   `json:"premium"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`

	// Orgowner   []string `json:"orgowner"`
	// Orgadmin   []string `json:"orgadmin"`
	// Orguser    []string `json:"orguser"`
	// Orgmanager []string `json:"orgmanager"`

	Sstamp string   `json:"sstamp"` // user security_stamp
	Device string   `json:"device"` // device uuid
	Scope  []string `json:"scope"`  // [ "api", "offline_access" ]
	Amr    []string `json:"amr"`    // [ "Application" ]
}

type Authenticator interface {
	EncodeJWT(claims jwt.Claims) (string, error)
}

type core struct {
	pubKey *rsa.PublicKey
	priKey *rsa.PrivateKey
}

func New() Authenticator {
	return &core{}
}

func (c core) EncodeJWT(claims jwt.Claims) (string, error) {
	tk := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	return tk.SignedString(c.priKey)
}

type JWTIssuer string

const (
	JWTIssuerLogin  = JWTIssuer("|login")
	JWTIssuerInvite = JWTIssuer("|invite")
)

func (c core) DecodeLogin(token string) (*LoginJwtClaims, error) {
	claims := &LoginJwtClaims{}
	err := c.decodeJWT(token, claims)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func (c core) decodeJWT(token string, claims jwt.Claims) error {
	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return c.pubKey, nil
	})
	if err != nil {
		return err
	}

	if !t.Valid {
		return errors.New("invalid token")
	}

	return nil
}
