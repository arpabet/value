package genval_impl

import (
	"reflect"
	"github.com/shvid/genval"
	"strconv"
)

type BoolVal bool


func (b BoolVal) Kind() genval.Kind {
	return genval.BoolVal
}

func (b BoolVal) Class() reflect.Type {
	return reflect.TypeOf(BoolVal(false))
}

func (b BoolVal) String() string {
	return strconv.FormatBool(bool(b))
}

func (b BoolVal) Pack(p genval.Packer) {
	p.PackBool(bool(b))
}

func (b BoolVal) Json() string {
	return b.String()
}

func (b BoolVal) Boolean() bool {
	return bool(b)
}


