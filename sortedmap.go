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
	"sort"
	"strings"
)

/**
	This is a sorted Map implementation with deterministic serialization process

	Serializes in MessagePack as Map with string index

	@author Alex Shvid
*/

type sortedMapEntry struct {
	key    string
	value  Value
}

type sortedMapValue []MapEntry
var sortedMapValueClass = reflect.TypeOf((*sortedMapValue)(nil)).Elem()

func Entry(key string, value Value) MapEntry {
	return &sortedMapEntry {
		key: key,
		value: value,
	}
}

func SortedMap(entries []MapEntry, sortedEntries bool) Map {
	t := sortedMapValue(entries)
	if !sortedEntries {
		sort.Sort(t)
	}
	return t
}

var emptySortedMap = sortedMapValue([]MapEntry{})

func EmptyMap() Map {
	return emptySortedMap
}

func (t *sortedMapEntry) Key() string {
	return t.key
}

func (t *sortedMapEntry) Value() Value {
	return t.value
}

func (t *sortedMapEntry) Equal(e MapEntry) bool {
	return t.key == e.Key() && Equal(t.value, e.Value())
}

func (t sortedMapValue) HashMap() map[string]Value {
	cache := make(map[string]Value)
	for _, entry := range t {
		cache[entry.Key()] = entry.Value()
	}
	return cache
}

func (t sortedMapValue) Entries() []MapEntry {
	return t
}

func (t sortedMapValue) Keys() []string {
	var keys []string
	for _, entry := range t {
		keys = append(keys, entry.Key())
	}
	return keys
}

func (t sortedMapValue) Values() []Value {
	var values []Value
	for _, entry := range t {
		values = append(values, entry.Value())
	}
	return values
}

func (t sortedMapValue) Len() int {
	return len(t)
}

func (t sortedMapValue) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t sortedMapValue) Less(i, j int) bool {
	return t[i].Key() < t[j].Key()
}

func (t sortedMapValue) Kind() Kind {
	return MAP
}

func (t sortedMapValue) Class() reflect.Type {
	return sortedMapValueClass
}

func (t sortedMapValue) Object() interface{} {
	return []MapEntry(t)
}

func (t sortedMapValue) String() string {
	var out strings.Builder
	t.PrintJSON(&out)
	return out.String()
}

func (t sortedMapValue) Pack(p Packer) {

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

func (t sortedMapValue) PrintJSON(out *strings.Builder) {

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

func (t sortedMapValue) MarshalJSON() ([]byte, error) {
	var out strings.Builder
	t.PrintJSON(&out)
	return []byte(out.String()), nil
}

func (t sortedMapValue) MarshalBinary() ([]byte, error) {
	buf := bytes.Buffer{}
	p := MessagePacker(&buf)
	t.Pack(p)
	return buf.Bytes(), p.Error()
}

func (t sortedMapValue) Equal(val Value) bool {
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

func (t sortedMapValue) Get(key string) (Value, bool) {
	n := len(t)
	i := sort.Search(n, func(i int) bool {
		return t[i].Key() >= key
	})
	if i == n {
		return nil, false
	} else if t[i].Key() == key {
		return t[i].Value(), true
	} else {
		return nil, false
	}
}

func (t sortedMapValue) GetBool(key string) Bool {
	value, _ := t.Get(key)
	if value != nil {
		if value.Kind() == BOOL {
			return value.(Bool)
		}
		return ParseBoolean(value.String())
	}
	return nil
}

func (t sortedMapValue) GetNumber(key string) Number {
	value, _ := t.Get(key)
	if value != nil {
		if value.Kind() == NUMBER {
			return value.(Number)
		}
		return ParseNumber(value.String())
	}
	return nil
}

func (t sortedMapValue) GetString(key string) String {
	value, _ := t.Get(key)
	if value != nil {
		if value.Kind() == STRING {
			return value.(String)
		}
		return ParseString(value.String())
	}
	return nil
}

func (t sortedMapValue) GetList(key string) List {
	value, _ := t.Get(key)
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

func (t sortedMapValue) GetMap(key string) Map {
	value, _ := t.Get(key)
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

func (t sortedMapValue) Insert(key string, value Value) Map {
	n := len(t)
	i := sort.Search(n, func(i int) bool {
		return t[i].Key() >= key
	})
	if i == n {
		return t.append(n, Entry(key, value))
	} else {
		return t.insertAt(i, n, Entry(key, value))
	}
}

func (t sortedMapValue) Put(key string, value Value) Map {
	n := len(t)
	i := sort.Search(n, func(i int) bool {
		return t[i].Key() >= key
	})
	if i == n {
		return t.append(n, Entry(key, value))
	} else if t[i].Key() == key {
		return t.replaceAt(i, n, Entry(key, value))
	} else {
		return t.insertAt(i, n, Entry(key, value))
	}
}

func (t sortedMapValue) Remove(key string) Map {
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

func (t sortedMapValue) append(n int, entry MapEntry) Map {
	if n == 0 {
		return sortedMapValue([]MapEntry{entry})
	} else if AllowFastAppends {   // fast appends are permitted w/o memory allocation
		return append(t, entry)
	} else {
		dst := make([]MapEntry, n+1)
		copy(dst, t)
		dst[n] = entry
		return sortedMapValue(dst)
	}
}

func (t sortedMapValue) replaceAt(i, n int, entry MapEntry) Map {
	dst := make([]MapEntry, n)
	copy(dst, t)
	dst[i] = entry
	return sortedMapValue(dst)
}

func (t sortedMapValue) insertAt(i, n int, entry MapEntry) Map {
	if i == 0 {
		dst := make([]MapEntry, n+1)
		copy(dst[1:], t)
		dst[0] = entry
		return sortedMapValue(dst)
	} else if i+1 == n {
		if AllowFastAppends {  // fast appends are permitted w/o memory allocation
			return append(t[:i], entry, t[i])
		} else {
			dst := make([]MapEntry, n+1)
			copy(dst, t[:i])
			dst[n-1] = entry
			dst[n] = t[i]
			return sortedMapValue(dst)
		}
	} else {
		dst := make([]MapEntry, n+1)
		copy(dst, t[:i])
		dst[i] = entry
		copy(dst[i+1:], t[i:])
		return sortedMapValue(dst)
	}
}

func (t sortedMapValue) removeAt(i, n int) Map {
	if i == 0 {
		return t[1:]
	} else if i+1 == n {
		return t[:i]
	} else if AllowFastAppends {
		return append(t[:i], t[i+1:]...)
	} else {
		dst := make([]MapEntry, n-1)
		copy(dst, t[:i])
		copy(dst[i:], t[i+1:])
		return sortedMapValue(dst)
	}
}

func (t sortedMapValue) Select(key string) []Value {
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

func (t sortedMapValue) InsertAll(key string, list []Value) Map {

	if len(list) == 0 {
		return t
	}

	var slice []MapEntry
	for _, value := range list {
		slice = append(slice, &sortedMapEntry{key, value})
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

func (t sortedMapValue) DeleteAll(key string) Map {
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

func (t sortedMapValue) appendSlice(n int, slice []MapEntry) Map {
	if n == 0 {
		return sortedMapValue(slice)
	} else if AllowFastAppends {  // fast appends are permitted w/o memory allocation
		return append(t, slice...)
	} else {
		dst := make([]MapEntry, n+len(slice))
		copy(dst, t)
		copy(dst[n:], slice)
		return sortedMapValue(dst)
	}
}

func (t sortedMapValue) insertSliceAt(i, n int, slice []MapEntry) Map {
	if i == 0 {
		if AllowFastAppends {
			return append(sortedMapValue(slice), t...)
		} else {
			m := len(slice)
			dst := make([]MapEntry, m+n)
			copy(dst, slice)
			copy(dst[m:], t)
			return sortedMapValue(dst)
		}
	} else {
		m := len(slice)
		dst := make([]MapEntry, n+m)
		copy(dst, t[:i])
		copy(dst[i:], slice)
		copy(dst[i+m:], t[i:])
		return sortedMapValue(dst)
	}
}

func (t sortedMapValue) removeSliceAt(i, cnt, n int) Map {
	if i == 0 {
		return t[cnt:]
	} else if i+cnt == n {
		return t[:i]
	} else if AllowFastAppends  {
		return append(t[:i], t[i+cnt:]...)
	} else {
		dst := make([]MapEntry, n-cnt)
		copy(dst, t[:i])
		copy(dst[i:], t[i+cnt:])
		return sortedMapValue(dst)
	}
}
