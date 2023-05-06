package example

import (
	"fmt"

	"github.com/siddhesh-tamhanekar/di/example/another/b"
)

type Db struct {
	dsn string
}

type UserServicer interface {
	process()
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

type UserHandler struct {
	ConfigRepo   *ConfigRepo
	UserServicer UserService
	B            *b.B2
	B1           *b.B1
}

var db Db

func main() {
	db = Db{
		dsn: "dsn://",
	}
	// b := true
	// // str := "ab"
	a := NewUserHandler("", []string{"a"})
	fmt.Println(a.B1)
}

// func NewUserRepo() (userRepo *UserRepo, err error) {
// 	if db.dsn != "123" {
// 		return nil, errors.New("db dsn incorrect")
// 	}
// 	return &UserRepo{
// 		DB: db,
// 	}, nil
// }
