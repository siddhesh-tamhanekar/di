package b

// Code generated by DI library. DO NOT EDIT.
// To generate file use <path_to_di>/di --path= --module=
var b1 = Bin()

func NewB2() (b2 B2) {

	b2 = B2{
		b1: b1,
	}
	return
}

func NewB1() B1 {

	return b1
}
