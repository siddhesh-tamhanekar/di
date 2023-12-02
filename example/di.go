//go:build exclude

package example

import (
	"github.com/siddhesh-tamhanekar/di"
	"github.com/siddhesh-tamhanekar/di/example/another/"
)

func NewUserHandler() UserHandler {
}

func build() {
	// di.Share(Db{}, db)
	// di.Share(b.B1{}, b.Bin())
	di.Singleton(Redis{}, NewRedis())

	di.Build(UserHandler{})
	// di.Build(b.B1{})
	di.Build(Simple{})
	di.Build(FooService{})
	di.Bind(Notification, map[string]any{
		"sms":   SmsNotification{},
		"gcm":   GcmNotification{},
		"email": another.EmailNotification{},
	})
	di.Singleton(Db{}, NewDb())

	di.Bind(UserServicer, UserService{})
	di.Bind(Writer, FileWriter{})
	di.BindEnv(UserServicer, TestUserService{}, "test")
}
