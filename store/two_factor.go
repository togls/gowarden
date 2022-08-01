package store

type TwoFactor interface {
	DeleteAllByUser(user string) error
}

type TwoFactorIncomplete interface {
	DeleteAllByUser(user string) error
}
