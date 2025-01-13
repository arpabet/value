/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package value

import (
	"bytes"
	"reflect"
	"strconv"
	"strings"
)

type immutableListEntry struct {
	key    string
	value  Value
}

type immutableListValue []Value

var immutableListValueClass = reflect.TypeOf((*immutableListValue)(nil)).Elem()

func EmptyImmutableList() List {
	return immutableListValue([]Value{})
}

func ImmutableList(list []Value) List {
	return immutableListValue(list)
}

func (t immutableListValue) Kind() Kind {
	return LIST
}

func (t immutableListValue) Class() reflect.Type {
	return immutableListValueClass
}

func (t immutableListValue) Object() interface{} {
	return []Value(t)
}

func (t immutableListValue) String() string {
	var out strings.Builder
	t.PrintJSON(&out)
	return out.String()
}

func (t immutableListValue) Items() []ListItem {
	var items []ListItem
	for key, value := range t {
		items = append(items, ImmutableItem(key, value))
	}
	return items
}

func (t immutableListValue) Entries() []MapEntry {
	var entries []MapEntry
	for key, value := range t {
		entries = append(entries, ImmutableEntry(strconv.Itoa(key), value))
	}
	return entries
}

func (t immutableListValue) Values() []Value {
	return t
}

func (t immutableListValue) Len() int {
	return len(t)
}

func (t immutableListValue) Pack(p Packer) {

	p.PackList(len(t))

	for _, e := range t {
		if e != nil {
			e.Pack(p)
		} else {
			p.PackNil()
		}
	}
}

func (t immutableListValue) PrintJSON(out *strings.Builder) {
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

func (t immutableListValue) MarshalJSON() ([]byte, error) {
	var out strings.Builder
	t.PrintJSON(&out)
	return []byte(out.String()), nil
}

func (t immutableListValue) MarshalBinary() ([]byte, error) {
	buf := bytes.Buffer{}
	p := MessagePacker(&buf)
	t.Pack(p)
	return buf.Bytes(), p.Error()
}

func (t immutableListValue) Equal(val Value) bool {
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

func (t immutableListValue) GetAt(i int) Value {
	if i >= 0 && i < len(t) {
		return t[i]
	}
	return Null
}

func (t immutableListValue) GetBoolAt(index int) Bool {
	value := t.GetAt(index)
	if value != Null {
		if value.Kind() == BOOL {
			return value.(Bool)
		}
		return ParseBoolean(value.String())
	}
	return False
}

func (t immutableListValue) GetNumberAt(index int) Number {
	value := t.GetAt(index)
	if value != Null {
		if value.Kind() == NUMBER {
			return value.(Number)
		}
		return ParseNumber(value.String())
	}
	return Zero
}

func (t immutableListValue) GetStringAt(index int) String {
	value := t.GetAt(index)
	if value != Null {
		if value.Kind() == STRING {
			return value.(String)
		}
		return ParseString(value.String())
	}
	return EmptyString
}

func (t immutableListValue) GetListAt(index int) List {
	value := t.GetAt(index)
	if value != Null {
		switch value.Kind() {
		case LIST:
			return value.(List)
		case MAP:
			return ImmutableList(value.(Map).Values())
		}
	}
	return EmptyImmutableList()
}

func (t immutableListValue) GetMapAt(index int) Map {
	value := t.GetAt(index)
	if value != Null {
		switch value.Kind() {
		case LIST:
			return ImmutableMap(value.(List).Entries(), false)
		case MAP:
			return value.(Map)
		}
	}
	return EmptyImmutableMap()
}

func (t immutableListValue) Append(val Value) List {
	if val == nil {
		val = Null
	}
	return t.append(len(t), val)
}

func (t immutableListValue) PutAt(i int, val Value) List {
	if val == nil {
		val = Null
	}
	n := len(t)
	if i >= 0 {
		if i == n {
			return t.append(n, val)
		} else {
			return t.putAt(i, n, val)
		}
	}
	return t
}

func (t immutableListValue) UpdateAt(int, Updater) bool {
	return false
}

func (t immutableListValue) InsertAt(i int, val Value) List {
	if val == nil {
		val = Null
	}
	if i >= 0 {
		n := len(t)
		if i < n {
			return t.insertAt(i, n, val)
		} else {
			return t.append(n, val)
		}
	}
	return t
}

func (t immutableListValue) RemoveAt(i int) List {
	n := len(t)
	if i >= 0 && i < n {
		return t.removeAt(i, n)
	}
	return t
}

func (t immutableListValue) append(n int, val Value) List {
	if n == 0 {
		return immutableListValue([]Value{val})
	} else {
		dst := make([]Value, n+1)
		copy(dst, t)
		dst[n] = val
		return immutableListValue(dst)
	}
}

func (t immutableListValue) putAt(i, n int, val Value) List {
	j := i+1
	if j < n {
		j = n
	}
	dst := make([]Value, j)
	copy(dst, t)
	dst[i] = val
	return immutableListValue(dst)
}

func (t immutableListValue) insertAt(i, n int, val Value) List {
	if i == 0 {
		dst := make([]Value, n+1)
		copy(dst[1:], t)
		dst[0] = val
		return immutableListValue(dst)
	} else if i+1 == n {
		dst := make([]Value, n+1)
		copy(dst, t[:i])
		dst[n-1] = val
		dst[n] = t[i]
		return immutableListValue(dst)
	} else {
		dst := make([]Value, n+1)
		copy(dst, t[:i])
		dst[i] = val
		copy(dst[i+1:], t[i:])
		return immutableListValue(dst)
	}
}

func (t immutableListValue) removeAt(i, n int) List {
	if i == 0 {
		return immutableListValue(t.copyOf(t[1:]))
	} else if i+1 == n {
		return immutableListValue(t.copyOf(t[:i]))
	} else {
		dst := make([]Value, n-1)
		copy(dst, t[:i])
		copy(dst[i:], t[i+1:])
		return immutableListValue(dst)
	}
}

func (t immutableListValue) Select(i int) []Value {
	val := t.GetAt(i)
	if val != Null {
		return []Value {val}
	}
	return []Value {}
}

func (t immutableListValue) InsertAll(i int, list []Value) List {

	if len(list) == 0 {
		return t
	}

	for k := range list {
		if list[k] == nil {
			list[k] = Null
		}
	}

	if i >= 0 {
		n := len(t)
		if i < n {
			return t.insertSliceAt(i, n, list)
		} else {
			return t.appendSlice(n, list)
		}
	}
	return t
}

func (t immutableListValue) DeleteAll(i int) List {
	return t.RemoveAt(i)
}

func (t immutableListValue) appendSlice(n int, slice []Value) List {
	if n == 0 {
		return immutableListValue(t.copyOf(slice))
	} else {
		dst := make([]Value, n+len(slice))
		copy(dst, t)
		copy(dst[n:], slice)
		return immutableListValue(dst)
	}
}

func (t immutableListValue) insertSliceAt(i, n int, slice []Value) List {
	if i == 0 {
		m := len(slice)
		dst := make([]Value, m+n)
		copy(dst, slice)
		copy(dst[m:], t)
		return immutableListValue(dst)
	} else {
		m := len(slice)
		dst := make([]Value, n+m)
		copy(dst, t[:i])
		copy(dst[i:], slice)
		copy(dst[i+m:], t[i:])
		return immutableListValue(dst)
	}
}

func (t immutableListValue) copyOf(src []Value) []Value {
	n := len(src)
	dst := make([]Value, n)
	copy(dst, src)
	return dst
}
