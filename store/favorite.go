package store

type Favorite interface {
	IsFavorite(cipher, user string) (bool, error)
	AddFavorite(cipher, user string) error
	DeleteAllByUser(user string) error
}
