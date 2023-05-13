package example

import (
	"errors"
	"fmt"

	"github.com/siddhesh-tamhanekar/di/example/another/b"
)

type Db struct {
	dsn string
}

type UserServicer interface {
	process()
}

type Writer interface {
	write(s string) bool
}

type Redis struct {
}

func NewRedis() Redis {
	return Redis{}
}

type UserRepo struct {
	Db    *Db
	Redis Redis
}

type ConfigRepo struct {
	names []string
}

type UserService struct {
	UserRepo UserRepo
}

func (u *UserService) process() {
}

func NewUserService() UserService {
	return UserService{}
}

type TestUserService struct {
	UserRepo UserRepo
}

func (u *TestUserService) process() {
}

type FileWriter struct{}

func (f FileWriter) write(s string) bool {
	return true
}

type UserHandler struct {
	ConfigRepo   *ConfigRepo
	UserServicer UserServicer
	w            Writer
	B            *b.B2
	B1           *b.B1
}

var db Db

func main() {
	db = Db{
		dsn: "dsn://",
	}
	a := NewUserHandler("", []string{"a"})
	fmt.Println(a.B1)
}

func NewUserRepo() (userRepo *UserRepo, err error) {
	if db.dsn != "123" {
		return nil, errors.New("db dsn incorrect")
	}
	return &UserRepo{
		Db: &db,
	}, nil
}
