/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */


package value

import (
	"reflect"
	"strings"
)

type nullValue struct{}

var Null = nullValue{}
var nullValueClass = reflect.TypeOf(Null)

func Nil() nullValue {
	return Null
}

func ParseNull(str string) Value {
	if str == "null" {
		return Null
	}
	return Utf8(str)
}

func (n nullValue) Kind() Kind {
	return NULL
}

func (n nullValue) Class() reflect.Type {
	return nullValueClass
}

func (n nullValue) Object() interface{} {
	return nil
}

func (n nullValue) String() string {
	return "null"
}

func (n nullValue) Pack(p Packer) {
	p.PackNil()
}

func (n nullValue) PrintJSON(out *strings.Builder) {
	out.WriteString(n.String())
}

func (n nullValue) MarshalJSON() ([]byte, error) {
	return []byte(n.String()), nil
}

func (n nullValue) MarshalBinary() ([]byte, error) {
	var m messageWriter
	return m.WriteNil(), nil
}

func (n nullValue) Equal(val Value) bool {
	return val == nil || val.Kind() == NULL
}
