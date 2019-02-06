package genval

import (
	"reflect"
	"strconv"
)

type BoolValue bool

func Boolean(b bool) Bool {
	return BoolValue(b)
}

func ParseBoolean(str string) BoolValue {
	b, _ := strconv.ParseBool(str)
	return BoolValue(b)
}

func (b BoolValue) Kind() Kind {
	return BOOL
}

func (b BoolValue) Class() reflect.Type {
	return reflect.TypeOf(BoolValue(false))
}

func (b BoolValue) String() string {
	return strconv.FormatBool(bool(b))
}

func (b BoolValue) Pack(p Packer) {
	p.PackBool(bool(b))
}

func (b BoolValue) Json() string {
	return b.String()
}

func (b BoolValue) Boolean() bool {
	return bool(b)
}


