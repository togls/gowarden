package crypto

import (
	"bytes"
	"crypto/sha256"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/pbkdf2"
)

func VerifyPassword(password string, salt, hash []byte, iter int) bool {
	pdb := pbkdf2.Key([]byte(password), salt, iter, 256/8, sha256.New)

	return bytes.Equal(pdb, hash)
}

func GeneratePassword(password string, salt []byte, iter int) []byte {
	return pbkdf2.Key([]byte(password), salt, iter, 256/8, sha256.New)
}

func GenerateBytes(n int) ([]byte, error) {
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)

	_, err := rd.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GenerateAlphanumString(n int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)

	for i := range b {
		b[i] = letters[rd.Intn(len(letters))]
	}

	return string(b), nil
}

func GenerateApiKey() (string, error) {
	return GenerateAlphanumString(30)
}

func GenerateUuid() (string, error) {
	rd, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	return rd.String(), nil
}
