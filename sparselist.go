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
	"strconv"
	"strings"
)

/**
	Position in list guarantees by it's index

	Serializes in MessagePack as Map with INT index

	@author Alex Shvid
*/

var FirstIndexKey = 0

type sparseListItem struct {
	key    int  // should be unsigned, but easy to maintain 'int'
	value  Value
}

type sparseListValue []ListItem
var sparseListValueClass = reflect.TypeOf((*sparseListValue)(nil)).Elem()

type sortableValues []ListItem

func (t sortableValues) Len() int {
	return len(t)
}

func (t sortableValues) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t sortableValues) Less(i, j int) bool {
	return t[i].Key() < t[j].Key()
}


func Item(key int, value Value) ListItem {
	return &sparseListItem{key, value}
}

var emptySparseList = sparseListValue([]ListItem{})

func EmptySparseList() List {
	return emptySparseList
}

func SparseListOf(list []Value) List {
	var items []ListItem
	for key, val := range list {
		if val != nil {
			items = append(items, Item(key, val))
		}
	}
	sort.Sort(sortableValues(items))
	return sparseListValue(items)
}

func SparseList(items []ListItem, sortedItems bool) List {
	if !sortedItems {
		sort.Sort(sortableValues(items))
	}
	return sparseListValue(items)
}

func SortedSparseList(items []ListItem) List {
	return sparseListValue(items)
}

func (t *sparseListItem) Key() int {
	return t.key
}

func (t *sparseListItem) Value() Value {
	return t.value
}

func (t *sparseListItem) Equal(e ListItem) bool {
	return t.key == e.Key() && Equal(t.value, e.Value())
}

func (t sparseListValue) Kind() Kind {
	return LIST
}

func (t sparseListValue) Class() reflect.Type {
	return sparseListValueClass
}

func (t sparseListValue) Object() interface{} {
	return []ListItem(t)
}

func (t sparseListValue) String() string {
	var out strings.Builder
	t.PrintJSON(&out)
	return out.String()
}

func (t sparseListValue) Items() []ListItem {
	return t
}

func (t sparseListValue) Entries() []MapEntry {
	var entries []MapEntry
	for _, item := range t {
		entries = append(entries, Entry(strconv.Itoa(item.Key()), item.Value()))
	}
	return entries
}

// ignore negative keys
func (t sparseListValue) Values() []Value {
	n := len(t)
	if n == 0 {
		return nil
	}
	maxKey := t[n-1].Key()
	values := make([]Value, maxKey+1)

	for _, item := range t {
		if item.Key() >= 0 {
			values[item.Key()] = item.Value()
		}
	}
	return values
}

func (t sparseListValue) Len() int {
	n := len(t)
	if n == 0 {
		return 0
	} else {
		maxKey := t[n-1].Key()
		return maxKey+1
	}
}

func (t sparseListValue) Pack(p Packer) {

	p.PackMap(len(t))

	for _, entry := range t {
		p.PackLong(int64(entry.Key()))
		value := entry.Value()
		if value != nil {
			value.Pack(p)
		} else {
			p.PackNil()
		}
	}

}

func (t sparseListValue) PrintJSON(out *strings.Builder) {

	out.WriteRune('{')
	for i, entry := range t {
		if i != 0 {
			out.WriteRune(',')
		}
		out.WriteRune(jsonQuote)
		out.WriteString(strconv.Itoa(entry.Key()))
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

func (t sparseListValue) MarshalJSON() ([]byte, error) {
	var out strings.Builder
	t.PrintJSON(&out)
	return []byte(out.String()), nil
}

func (t sparseListValue) MarshalBinary() ([]byte, error) {
	buf := bytes.Buffer{}
	p := MessagePacker(&buf)
	t.Pack(p)
	return buf.Bytes(), p.Error()
}

func (t sparseListValue) Equal(val Value) bool {
	if val == nil || val.Kind() != LIST {
		return false
	}
	o := val.(List)
	if t.Len() != o.Len() {
		return false
	}
	// entries are sorted
	other := o.Items()
	for i, item := range t {
		if !item.Equal(other[i]) {
			return false
		}
	}
	return true
}

func (t sparseListValue) GetAt(key int) Value {
	n := len(t)
	i := sort.Search(n, func(i int) bool {
		return t[i].Key() >= key
	})
	if i == n {
		return nil
	} else if t[i].Key() == key {
		return t[i].Value()
	} else {
		return nil
	}
}

func (t sparseListValue) GetBoolAt(index int) Bool {
	value := t.GetAt(index)
	if value != nil {
		if value.Kind() == BOOL {
			return value.(Bool)
		}
		return ParseBoolean(value.String())
	}
	return nil
}

func (t sparseListValue) GetNumberAt(index int) Number {
	value := t.GetAt(index)
	if value != nil {
		if value.Kind() == NUMBER {
			return value.(Number)
		}
		return ParseNumber(value.String())
	}
	return nil
}

func (t sparseListValue) GetStringAt(index int) String {
	value := t.GetAt(index)
	if value != nil {
		if value.Kind() == STRING {
			return value.(String)
		}
		return ParseString(value.String())
	}
	return nil
}

func (t sparseListValue) GetListAt(index int) List {
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

func (t sparseListValue) GetMapAt(index int) Map {
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

func (t sparseListValue) Append(value Value) List {
	n := len(t)
	if n == 0 {
		return t.append(n, Item(FirstIndexKey, value))
	} else {
		maxKey := t[n-1].Key()
		return t.append(n, Item(maxKey+1, value))
	}
}

func (t sparseListValue) PutAt(key int, value Value) List {
	n := len(t)
	i := sort.Search(n, func(i int) bool {
		return t[i].Key() >= key
	})
	if i == n {
		return t.append(n, Item(key, value))
	} else if t[i].Key() == key {
		return t.replaceAt(i, n, Item(key, value))
	} else {
		return t.insertAt(i, n, Item(key, value))
	}
}

func (t sparseListValue) InsertAt(key int, value Value) List {
	n := len(t)
	i := sort.Search(n, func(i int) bool {
		return t[i].Key() >= key
	})
	if i == n {
		return t.append(n, Item(key, value))
	} else {
		return t.insertAt(i, n, Item(key, value))
	}
}

func (t sparseListValue) RemoveAt(key int) List {
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

func (t sparseListValue) append(n int, item ListItem) List {
	if n == 0 {
		return sparseListValue([]ListItem{item})
	} else if AllowFastAppends {  // fast appends are permitted w/o memory allocation
		return append(t, item)
	} else {
		dst := make([]ListItem, n+1)
		copy(dst, t)
		dst[n] = item
		return sparseListValue(dst)
	}
}

func (t sparseListValue) replaceAt(i, n int, item ListItem) List {
	dst := make([]ListItem, n)
	copy(dst, t)
	dst[i] = item
	return sparseListValue(dst)
}

func (t sparseListValue) insertAt(i, n int, item ListItem) List {
	if i == 0 {
		dst := make([]ListItem, n+1)
		copy(dst[1:], t)
		dst[0] = item
		return sparseListValue(dst)
	} else if i+1 == n {
		if AllowFastAppends {  // fast appends are permitted w/o memory allocation
			return append(t[:i], item, t[i])
		} else {
			dst := make([]ListItem, n+1)
			copy(dst, t[:i])
			dst[n-1] = item
			dst[n] = t[i]
			return sparseListValue(dst)
		}
	} else {
		dst := make([]ListItem, n+1)
		copy(dst, t[:i])
		dst[i] = item
		copy(dst[i+1:], t[i:])
		return sparseListValue(dst)
	}
}

func (t sparseListValue) removeAt(i, n int) List {
	if i == 0 {
		return t[1:]
	} else if i+1 == n {
		return t[:i]
	} else if AllowFastAppends {
		return append(t[:i], t[i+1:]...)
	} else {
		dst := make([]ListItem, n-1)
		copy(dst, t[:i])
		copy(dst[i:], t[i+1:])
		return sparseListValue(dst)
	}
}

func (t sparseListValue) Select(key int) []Value {
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

func (t sparseListValue) InsertAll(key int, list []Value) List {

	if len(list) == 0 {
		return t
	}

	var slice []ListItem
	for _, value := range list {
		slice = append(slice, &sparseListItem{key, value})
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

func (t sparseListValue) DeleteAll(key int) List {
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

func (t sparseListValue) appendSlice(n int, slice []ListItem) List {
	if n == 0 {
		return sparseListValue(slice)
	} else if AllowFastAppends {  // fast appends are permitted w/o memory allocation
		return append(t, slice...)
	} else {
		dst := make([]ListItem, n+len(slice))
		copy(dst, t)
		copy(dst[n:], slice)
		return sparseListValue(dst)
	}
}

func (t sparseListValue) insertSliceAt(i, n int, slice []ListItem) List {
	if i == 0 {
		if AllowFastAppends {
			return append(sparseListValue(slice), t...)
		} else {
			m := len(slice)
			dst := make([]ListItem, m+n)
			copy(dst, slice)
			copy(dst[m:], t)
			return sparseListValue(dst)
		}
	} else {
		m := len(slice)
		dst := make([]ListItem, n+m)
		copy(dst, t[:i])
		copy(dst[i:], slice)
		copy(dst[i+m:], t[i:])
		return sparseListValue(dst)
	}
}

func (t sparseListValue) removeSliceAt(i, cnt, n int) List {
	if i == 0 {
		return t[cnt:]
	} else if i+cnt == n {
		return t[:i]
	} else if AllowFastAppends  {
		return append(t[:i], t[i+cnt:]...)
	} else {
		dst := make([]ListItem, n-cnt)
		copy(dst, t[:i])
		copy(dst[i:], t[i+cnt:])
		return sparseListValue(dst)
	}
}
