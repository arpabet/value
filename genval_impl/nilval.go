package genval_impl

import (
	"github.com/shvid/genval"
	"reflect"
)

type NilVal int

func (n NilVal) Kind() genval.Kind {
	return genval.NilVal
}

func (n NilVal) Class() reflect.Type {
	return reflect.TypeOf(NilVal(0))
}

func (n NilVal) String() string {
	return "nil"
}

func (n NilVal) Pack(p genval.Packer) {
	p.PackNil()
}

func (n NilVal) Json() string {
	return "null"
}

