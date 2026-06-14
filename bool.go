/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */


package value


import (
	"reflect"
	"strconv"
	"strings"
)

type boolValue bool

var True = boolValue(true)
var False = boolValue(false)
var boolValueClass = reflect.TypeOf(False)

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
	return boolValueClass
}

func (b boolValue) Object() interface{} {
	return bool(b)
}

func (b boolValue) String() string {
	return strconv.FormatBool(bool(b))
}

func (b boolValue) Boolean() bool {
	return bool(b)
}

func (b boolValue) Pack(p Packer) {
	p.PackBool(bool(b))
}

func (b boolValue) PrintJSON(out *strings.Builder) {
	out.WriteString(b.String())
}

func (b boolValue) MarshalJSON() ([]byte, error) {
	return []byte(b.String()), nil
}

func (b boolValue) MarshalBinary() ([]byte, error) {
	var m messageWriter
	return m.WriteBool(bool(b)), nil
}

func (b boolValue) Equal(val Value) bool {
	if val == nil || val.Kind() != BOOL {
		return false
	}
	o := val.(Bool)
	return b.Boolean() == o.Boolean()
}
