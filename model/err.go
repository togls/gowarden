package model

const ErrNotFound = storeErr("item not found")

type storeErr string

func (e storeErr) Error() string {
	return string(e)
}
