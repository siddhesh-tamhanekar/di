package other

// Code generated by DI library. DO NOT EDIT.
// To generate file use <path_to_di>/di --path= --module=
import "github.com/siddhesh-tamhanekar/di/example/another/b"

func NewOther() (other Other) {
	b1 := b.NewB1()
	other = Other{
		b: b1,
	}
	return
}