package b

func bin() {

}

func Bin() B1 {
	return B1{
		name: "123",
	}
}

type B1 struct {
	name string
}

type B2 struct {
	b1 B1
}
