package genval

import (
	"reflect"
	"strconv"
)

type boolValue bool

func (b boolValue) Equal(val Value) bool {
	if val == nil || val.Kind() != BOOL {
		return false
	}
	o := val.(boolValue)
	return b == o
}

func Boolean(b bool) Bool {
	return boolValue(b)
}

func ParseBoolean(str string) boolValue {
	b, _ := strconv.ParseBool(str)
	return boolValue(b)
}

func (b boolValue) Kind() Kind {
	return BOOL
}

func (b boolValue) Class() reflect.Type {
	return reflect.TypeOf(boolValue(false))
}

func (b boolValue) String() string {
	return strconv.FormatBool(bool(b))
}

func (b boolValue) Pack(p Packer) {
	p.PackBool(bool(b))
}

func (b boolValue) Json() string {
	return b.String()
}

func (b boolValue) Boolean() bool {
	return bool(b)
}


