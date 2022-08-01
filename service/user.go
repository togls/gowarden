package service

import "github.com/togls/gowarden/model"

type User struct {
	data *model.User
}

func (u User) CheckValidPassword(pw string) bool {
	return false
}
