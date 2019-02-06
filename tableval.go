package genval

import "reflect"

//
// All indexes start from 1, but it is configurable
//

var FirstIndex = 1

type tableValue struct {
	tt 		TableType
}

func List() *tableValue {
	return &tableValue{tt: LIST}
}

func Map() *tableValue {
	return &tableValue{tt: MAP}
}

func (t tableValue) Kind() Kind {
	return TABLE
}

func (t tableValue) Class() reflect.Type {
	return reflect.TypeOf((*tableValue)(nil)).Elem()
}

func (t tableValue) String() string {
	return "{}"
}

func (t tableValue) Pack(p Packer) {
	p.PackNil()
}

func (t tableValue) Json() string {
	return "{}"
}
