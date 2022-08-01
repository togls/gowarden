package store

import "github.com/togls/gowarden/model"

type Collection interface {
	Find(filter *model.CollectionFilter) (model.CollectionList, error)

	FindByUuid(uuid string) (*model.Collection, error)
	FindByCipherAndOrg(cipher, org string) (*model.Collection, error)
	FindByCollectionUser(collection, user string) (*model.Collection, error)
	FindByCollectionOrg(collection, org string) (*model.Collection, error)

	Save(c *model.Collection) error
	Delete(uuid string) error

	// CipherCollection

	FindCollectionIds(cipher, user string) ([]string, error)
	SaveCipher(collectionIDs []string, cipher string) error
	DeleteCipher(collectionIDs []string, cipher string) error

	// UserCollection

	SaveUser(collectionIDs []string, user string, readOnly, hidePasswords bool) error
	DeleteUser(collectionIDs []string, user string) error

	CollectionWriteable(collection, user string) (bool, error)
}

type UserCollection interface {
	Save(collection, user string, readOnly, hidePasswords bool) error

	Find(filter *model.UCFilter) (model.UCList, error)

	FindByCollectionUser(collection, user string) (*model.UserCollection, error)
	FindByUserCipher(user, cipher string) (*model.UserCollection, error)

	DeleteAllByCollection(collection string) error
	DeleteAllByUserAndOrg(user, org string) error
	DeleteByUserCollection(collection, user string) error
}
