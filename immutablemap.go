/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package value

import (
	"bytes"
	"reflect"
	"sort"
	"strings"
)

/**
This is an immutable sorted Map implementation with deterministic serialization process

Serializes in MessagePack as Map with string index
*/

type immutableMapValue []MapEntry

var immutableMapValueClass = reflect.TypeOf((*immutableMapValue)(nil)).Elem()

func EmptyImmutableMap() Map {
	return immutableMapValue([]MapEntry{})
}

func ImmutableMapOf(src map[string]Value) Map {
	entries := make([]MapEntry, len(src))
	var i int
	for key, value := range src {
		entries[i] = ImmutableEntry(key, value)
		i++
	}
	t := immutableMapValue(entries)
	sort.Sort(t)
	return t
}

func ImmutableMap(entries []MapEntry, sorted bool) Map {
	t := immutableMapValue(entries)
	if !sorted {
		sort.Sort(t)
	}
	return t
}

func ImmutableMapCopyOf(other Map) Map {
	t := immutableMapValue(mapEntryCopyOf(other.Entries()))
	return t
}

func (t immutableMapValue) HashMap() map[string]Value {
	cache := make(map[string]Value)
	for _, entry := range t {
		cache[entry.Key()] = entry.Value()
	}
	return cache
}

func (t immutableMapValue) Entries() []MapEntry {
	return t
}

func (t immutableMapValue) Keys() []string {
	var keys []string
	for _, entry := range t {
		keys = append(keys, entry.Key())
	}
	return keys
}

func (t immutableMapValue) Values() []Value {
	var values []Value
	for _, entry := range t {
		values = append(values, entry.Value())
	}
	return values
}

func (t immutableMapValue) Len() int {
	return len(t)
}

func (t immutableMapValue) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t immutableMapValue) Less(i, j int) bool {
	return t[i].Key() < t[j].Key()
}

func (t immutableMapValue) Kind() Kind {
	return MAP
}

func (t immutableMapValue) Class() reflect.Type {
	return immutableMapValueClass
}

func (t immutableMapValue) Object() interface{} {
	return []MapEntry(t)
}

func (t immutableMapValue) String() string {
	var out strings.Builder
	t.PrintJSON(&out)
	return out.String()
}

func (t immutableMapValue) Pack(p Packer) {

	p.PackMap(len(t))

	for _, entry := range t {
		p.PackStr(entry.Key())
		value := entry.Value()
		if value != nil {
			value.Pack(p)
		} else {
			p.PackNil()
		}
	}

}

func (t immutableMapValue) PrintJSON(out *strings.Builder) {

	out.WriteRune('{')
	for i, entry := range t {
		if i != 0 {
			out.WriteRune(',')
		}
		out.WriteRune(jsonQuote)
		out.WriteString(entry.Key())
		out.WriteRune(jsonQuote)

		out.WriteString(": ")
		value := entry.Value()
		if value != nil {
			value.PrintJSON(out)
		} else {
			out.WriteString("null")
		}
	}
	out.WriteRune('}')
}

func (t immutableMapValue) MarshalJSON() ([]byte, error) {
	var out strings.Builder
	t.PrintJSON(&out)
	return []byte(out.String()), nil
}

func (t immutableMapValue) MarshalBinary() ([]byte, error) {
	buf := bytes.Buffer{}
	p := MessagePacker(&buf)
	t.Pack(p)
	return buf.Bytes(), p.Error()
}

func (t immutableMapValue) Equal(val Value) bool {
	if val == nil || val.Kind() != MAP {
		return false
	}
	o := val.(Map)
	if t.Len() != o.Len() {
		return false
	}
	// entries are sorted
	other := o.Entries()
	for i, entry := range t {
		if !entry.Equal(other[i]) {
			return false
		}
	}
	return true
}

func (t immutableMapValue) Get(key string) Value {
	n := len(t)
	i := sort.Search(n, func(i int) bool {
		return t[i].Key() >= key
	})
	if i == n {
		return Null
	} else if t[i].Key() == key {
		return t[i].Value()
	} else {
		return Null
	}
}

func (t immutableMapValue) GetBool(key string) Bool {
	value := t.Get(key)
	if value != Null {
		if value.Kind() == BOOL {
			return value.(Bool)
		}
		return ParseBoolean(value.String())
	}
	return False
}

func (t immutableMapValue) GetNumber(key string) Number {
	value := t.Get(key)
	if value != Null {
		if value.Kind() == NUMBER {
			return value.(Number)
		}
		return ParseNumber(value.String())
	}
	return Zero
}

func (t immutableMapValue) GetString(key string) String {
	value := t.Get(key)
	if value != Null {
		if value.Kind() == STRING {
			return value.(String)
		}
		return ParseString(value.String())
	}
	return EmptyString
}

func (t immutableMapValue) GetList(key string) List {
	value := t.Get(key)
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

func (t immutableMapValue) GetMap(key string) Map {
	value := t.Get(key)
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

func (t immutableMapValue) Insert(key string, value Value) Map {
	if value == nil {
		value = Null
	}
	n := len(t)
	i := sort.Search(n, func(i int) bool {
		return t[i].Key() >= key
	})
	if i == n {
		return t.append(n, ImmutableEntry(key, value))
	} else {
		return t.insertAt(i, n, ImmutableEntry(key, value))
	}
}

func (t immutableMapValue) Put(key string, value Value) Map {
	if value == nil {
		value = Null
	}
	n := len(t)
	i := sort.Search(n, func(i int) bool {
		return t[i].Key() >= key
	})
	if i == n {
		return t.append(n, ImmutableEntry(key, value))
	} else if t[i].Key() == key {
		return t.replaceAt(i, n, ImmutableEntry(key, value))
	} else {
		return t.insertAt(i, n, ImmutableEntry(key, value))
	}
}

func (t immutableMapValue) Update(key string, updater Updater) bool {
	return false
}

func (t immutableMapValue) Remove(key string) Map {
	n := len(t)
	i := sort.Search(n, func(i int) bool {
		return t[i].Key() >= key
	})
	if i == n {
		return t
	} else if t[i].Key() == key {
		return t.removeAt(i, n)
	} else {
		return t
	}
}

func (t immutableMapValue) append(n int, entry MapEntry) Map {
	if n == 0 {
		return immutableMapValue([]MapEntry{entry})
	} else {
		dst := make([]MapEntry, n+1)
		copy(dst, t)
		dst[n] = entry
		return immutableMapValue(dst)
	}
}

func (t immutableMapValue) replaceAt(i, n int, entry MapEntry) Map {
	dst := make([]MapEntry, n)
	copy(dst, t)
	dst[i] = entry
	return immutableMapValue(dst)
}

func (t immutableMapValue) insertAt(i, n int, entry MapEntry) Map {
	if i == 0 {
		dst := make([]MapEntry, n+1)
		copy(dst[1:], t)
		dst[0] = entry
		return immutableMapValue(dst)
	} else if i+1 == n {
		dst := make([]MapEntry, n+1)
		copy(dst, t[:i])
		dst[n-1] = entry
		dst[n] = t[i]
		return immutableMapValue(dst)
	} else {
		dst := make([]MapEntry, n+1)
		copy(dst, t[:i])
		dst[i] = entry
		copy(dst[i+1:], t[i:])
		return immutableMapValue(dst)
	}
}

func (t immutableMapValue) removeAt(i, n int) Map {
	if i == 0 {
		return t[1:]
	} else if i+1 == n {
		return t[:i]
	}  else {
		dst := make([]MapEntry, n-1)
		copy(dst, t[:i])
		copy(dst[i:], t[i+1:])
		return immutableMapValue(dst)
	}
}

func (t immutableMapValue) Select(key string) []Value {
	n := len(t)
	i := sort.Search(n, func(i int) bool {
		return t[i].Key() >= key
	})
	var list []Value
	for j := i; j < n && t[j].Key() == key; j++ {
		list = append(list, t[j].Value())
	}
	return list
}

func (t immutableMapValue) InsertAll(key string, list []Value) Map {

	if len(list) == 0 {
		return t
	}

	for k := range list {
		if list[k] == nil {
			list[k] = Null
		}
	}

	slice := make([]MapEntry, len(list))
	for i, value := range list {
		slice[i] = &immutableMapEntry{key, value}
	}

	n := len(t)
	i := sort.Search(n, func(i int) bool {
		return t[i].Key() >= key
	})
	if i == n {
		return t.appendSlice(n, slice)
	} else {
		return t.insertSliceAt(i, n, slice)
	}
}

func (t immutableMapValue) DeleteAll(key string) Map {
	n := len(t)
	i := sort.Search(n, func(i int) bool {
		return t[i].Key() >= key
	})
	if i == n {
		return t
	}
	cnt := 0
	for j := i; j < n && t[j].Key() == key; j++ {
		cnt++
	}
	return t.removeSliceAt(i, cnt, n)
}

func (t immutableMapValue) appendSlice(n int, slice []MapEntry) Map {
	if n == 0 {
		return immutableMapValue(mapEntryCopyOf(slice))
	} else {
		dst := make([]MapEntry, n+len(slice))
		copy(dst, t)
		copy(dst[n:], slice)
		return immutableMapValue(dst)
	}
}

func (t immutableMapValue) insertSliceAt(i, n int, slice []MapEntry) Map {
	if i == 0 {
		m := len(slice)
		dst := make([]MapEntry, m+n)
		copy(dst, slice)
		copy(dst[m:], t)
		return immutableMapValue(dst)
	} else {
		m := len(slice)
		dst := make([]MapEntry, n+m)
		copy(dst, t[:i])
		copy(dst[i:], slice)
		copy(dst[i+m:], t[i:])
		return immutableMapValue(dst)
	}
}

func (t immutableMapValue) removeSliceAt(i, cnt, n int) Map {
	if i == 0 {
		return immutableMapValue(mapEntryCopyOf(t[cnt:]))
	} else if i+cnt == n {
		return immutableMapValue(mapEntryCopyOf(t[:i]))
	} else {
		dst := make([]MapEntry, n-cnt)
		copy(dst, t[:i])
		copy(dst[i:], t[i+cnt:])
		return immutableMapValue(dst)
	}
}

