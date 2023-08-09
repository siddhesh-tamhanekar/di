package example

// Code generated by DI library. DO NOT EDIT.
// To generate file use <path_to_di>/di --path= --module=
import (
	"github.com/siddhesh-tamhanekar/di/example/another/b"
)

var redis = NewRedis()

func NewUserHandler(names []string) (userHandler UserHandler) {
	configRepo := ConfigRepo{
		names: names,
	}

	userRepo := UserRepo{
		Db:    &db,
		Redis: redis,
	}
	userService := UserService{
		UserRepo: userRepo,
	}
	fileWriter := FileWriter{}
	b2 := b.NewB2()
	b1 := b.NewB1()
	userHandler = UserHandler{
		ConfigRepo:   &configRepo,
		UserServicer: &userService,
		w:            fileWriter,
		B:            &b2,
		B1:           &b1,
	}
	return
}

func NewUserServicer() (userService UserServicer) {

	userRepo := UserRepo{
		Db:    &db,
		Redis: redis,
	}
	userService = &UserService{
		UserRepo: userRepo,
	}
	return
}

func NewWriter() (fileWriter Writer) {
	fileWriter = FileWriter{}
	return
}
