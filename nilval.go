package genval

import (
	"reflect"
)

type NilValue int

func Nil() NilValue {
	return NilValue(0)
}

func (n NilValue) Kind() Kind {
	return NIL
}

func (n NilValue) Class() reflect.Type {
	return reflect.TypeOf(NilValue(0))
}

func (n NilValue) String() string {
	return "nil"
}

func (n NilValue) Pack(p Packer) {
	p.PackNil()
}

func (n NilValue) Json() string {
	return "null"
}

