package example

import (
	"net/http"

	"github.com/siddhesh-tamhanekar/di/example/another/b"
	"gorm.io/gorm"
)

type Notification interface {
	Send(msg string) bool
}

type SmsNotification struct {
}

func (s *SmsNotification) Send(msg string) bool {
	return true
}

type GcmNotification struct {
	redis Redis
}

func (s *GcmNotification) Send(msg string) bool {
	return true
}

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

func NewRedis() (*Redis, error) {
	return &Redis{}, nil
}

func NewDb() *Db {
	return &Db{}
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

// func NewUserService() UserService {
// 	return UserService{}
// }

type TestUserService struct {
	UserRepo UserRepo
}

func (u *TestUserService) process() {
}

type FileWriter struct {
	fname string
}

func (f FileWriter) write(s string) bool {
	return true
}

type UserHandler struct {
	ConfigRepo   *ConfigRepo
	UserServicer UserServicer
	w            Writer
	B            b.B2
	B1           *b.B122
}

var db *gorm.DB

type Foo struct {
	Name string
	B1   b.B1
}

// func NewFoo(name string) (Foo, error) {
// 	return Foo{name}, nil
// }

type Simple struct {
	age   []int
	marks map[string]int
	C     *http.Request
	Foo   *Foo
	W     Writer
}

func main() {

	// var s []UserServicer
	// us, _ := NewwUserServicer()
	// s = append(s, us)
	// fmt.Println(s)
}

type Logger struct {
	file string
}
type Infra struct {
	Db     *gorm.DB
	Logger Logger
}

type Bar struct {
	Baz    string
	Logger Logger
}
type FooService struct {
	Infra
	Bar Bar
}

// func NewUserRepo() (userRepo *UserRepo, err error) {
// 	if db.dsn != "123" {
// 		return nil, errors.New("db dsn incorrect")
// 	}
// 	return &UserRepo{
// 		Db: &db,
// 	}, nil
// }
