package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"os"
)

func WritePrivateKeyToPem(key *rsa.PrivateKey, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	return pem.Encode(file, block)
}

func ReadPrivateKeyFromPem(path string) (*rsa.PrivateKey, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	keyByte, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyByte)

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func WritePublicKeyToPem(key *rsa.PublicKey, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	block := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(key),
	}

	return pem.Encode(file, block)
}

func ReadPublicKeyFromPem(path string) (*rsa.PublicKey, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	keyByte, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyByte)

	return x509.ParsePKCS1PublicKey(block.Bytes)
}
