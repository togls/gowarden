package service

type Collection interface {
	SaveCipher(collection, cipher string) error
	SaveUser(collection, user string, readOnly, hidePasswords bool) error
	CollectionWriteable(collection, user string) (bool, error)
	CipherAccess(cipher, user string) (readOnly, hidePassword bool, err error)
}
