package genval

import (
	"reflect"
)

type nilValue int

func Nil() nilValue {
	return nilValue(0)
}

func (n nilValue) Kind() Kind {
	return NIL
}

func (n nilValue) Class() reflect.Type {
	return reflect.TypeOf(nilValue(0))
}

func (n nilValue) String() string {
	return "nil"
}

func (n nilValue) Pack(p Packer) {
	p.PackNil()
}

func (n nilValue) Json() string {
	return "null"
}

