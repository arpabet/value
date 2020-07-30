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
	"bytes"
	"reflect"
	"strconv"
	"strings"
)

/**
	Position in list is important, that guarantees the order

	Serializes in MessagePack as List

	@author Alex Shvid
*/

type solidListValue []Value

var emptySolidList = solidListValue([]Value{})

func EmptyList() List {
	return emptySolidList
}

func SolidList(list []Value) List {
	return solidListValue(list)
}

func Sequence(values... Value) List {
	return solidListValue(values)
}

func (t solidListValue) Kind() Kind {
	return LIST
}

func (t solidListValue) Class() reflect.Type {
	return reflect.TypeOf((*solidListValue)(nil)).Elem()
}

func (t solidListValue) Object() interface{} {
	return []Value(t)
}

func (t solidListValue) String() string {
	var out strings.Builder
	t.PrintJSON(&out)
	return out.String()
}

func (t solidListValue) Items() []ListItem {
	var items []ListItem
	for key, value := range t {
		items = append(items, Item(key, value))
	}
	return items
}

func (t solidListValue) Entries() []MapEntry {
	var entries []MapEntry
	for key, value := range t {
		entries = append(entries, Entry(strconv.Itoa(key), value))
	}
	return entries
}

func (t solidListValue) Values() []Value {
	return t
}

func (t solidListValue) Len() int {
	return len(t)
}

func (t solidListValue) Pack(p Packer) {

	p.PackList(len(t))

	for _, e := range t {
		if e != nil {
			e.Pack(p)
		} else {
			p.PackNil()
		}
	}
}

func (t solidListValue) PrintJSON(out *strings.Builder) {
	out.WriteRune('[')
	for i, e := range t {
		if i != 0 {
			out.WriteRune(',')
		}
		if e != nil {
			e.PrintJSON(out)
		} else {
			out.WriteString("null")
		}
	}
	out.WriteRune(']')
}

func (t solidListValue) MarshalJSON() ([]byte, error) {
	var out strings.Builder
	t.PrintJSON(&out)
	return []byte(out.String()), nil
}

func (t solidListValue) MarshalBinary() ([]byte, error) {
	buf := bytes.Buffer{}
	p := MessagePacker(&buf)
	t.Pack(p)
	return buf.Bytes(), p.Error()
}

func (t solidListValue) Equal(val Value) bool {
	if val == nil || val.Kind() != LIST {
		return false
	}
	o := val.(List)
	if t.Len() != o.Len() {
		return false
	}
	for i, item := range t {
		if !Equal(item, o.GetAt(i)) {
			return false
		}
	}
	return true
}

func (t solidListValue) GetAt(i int) Value {
	if i >= 0 && i < len(t) {
		return t[i]
	}
	return nil
}

func (t solidListValue) GetBoolAt(index int) Bool {
	value := t.GetAt(index)
	if value != nil {
		if value.Kind() == BOOL {
			return value.(Bool)
		}
		return ParseBoolean(value.String())
	}
	return nil
}

func (t solidListValue) GetNumberAt(index int) Number {
	value := t.GetAt(index)
	if value != nil {
		if value.Kind() == NUMBER {
			return value.(Number)
		}
		return ParseNumber(value.String())
	}
	return nil
}

func (t solidListValue) GetStringAt(index int) String {
	value := t.GetAt(index)
	if value != nil {
		if value.Kind() == STRING {
			return value.(String)
		}
		return ParseString(value.String())
	}
	return nil
}

func (t solidListValue) GetListAt(index int) List {
	value := t.GetAt(index)
	if value != nil {
		switch value.Kind() {
		case LIST:
			return value.(List)
		case MAP:
			return SolidList(value.(Map).Values())
		}
	}
	return nil
}

func (t solidListValue) GetMapAt(index int) Map {
	value := t.GetAt(index)
	if value != nil {
		switch value.Kind() {
		case LIST:
			return SortedMap(value.(List).Entries(), false)
		case MAP:
			return value.(Map)
		}
	}
	return nil
}

func (t solidListValue) Append(v Value) List {
	return append(t, v)
}

func (t solidListValue) PutAt(i int, value Value) List {
	n := len(t)
	if i >= 0 {
		if i == n {
			return append(t, value)
		}
		j := i+1
		if j < n {
			j = n
		}
		dst := make([]Value, j)
		copy(dst, t)
		dst[i] = value
		return solidListValue(dst)
	}
	return t
}

func (t solidListValue) InsertAt(i int, v Value) List {
	if i >= 0 {
		if i < len(t) {
			return append(append(t[:i], v), t[i+1:]...)
		} else {
			return append(t, v)
		}
	}
	return t
}

func (t solidListValue) RemoveAt(i int) List {
	if i >= 0 && i < len(t) {
		return append(t[:i], t[i+1:]...)
	}
	return t
}