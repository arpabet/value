/*
 *
 * Copyright 2020-present Arpabet, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package value

import (
	"reflect"
	"strconv"
	"strings"
)

type boolValue bool

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
