//go:build exclude

package main

import (
	di "di/lib"
)

func NewUserHandler() UserHandler {
}

func build() {
	di.Share(Db{}, db)
	di.Share(Redis{}, NewRedis())
	// di.Share(b.B1{}, b.Bin())

	di.Build(UserHandler{})

	di.Bind(UserServicer, UserService{})
}
