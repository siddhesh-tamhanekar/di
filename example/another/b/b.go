package b

func bin() {

}

type B1 struct {
	name string
}

func NewB1() *B1 {
	return &B1{
		name: "123",
	}
}

type B122 struct {
	Name string
}

type B2 struct {
	B11 B122
}
