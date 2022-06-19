/**
    Copyright (c) 2020-2022 Arpabet, Inc.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in
	all copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
	THE SOFTWARE.
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
