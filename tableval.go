package genval

import (
	"reflect"
	"strconv"
	"sort"
	"fmt"
	"strings"
)

//
// All indexes start from 1 like in LUA language
//

var FirstIndex = 1
var InitTableSize = 16

type keyType int

const (
	InvalidKey keyType = iota
	INDEX
	KEY
)

type opCode int

const (
	InvalidOp opCode = iota
	PUT
	REMOVE
)

type tableKey struct {
	typ   	keyType
	index 	int
	key   	string
}

func (l tableKey) Compare(r tableKey) int {

	if l.typ < r.typ {
		return -1
	} else if l.typ > r.typ {
		return 1
	}

	switch l.typ {

	case INDEX:
		if l.index < r.index {
			return -1
		} else if l.index > r.index {
			return 1
		}

	case KEY:
		if l.key < r.key {
			return -1
		} else if l.key > r.key {
			return 1
		}
	}

	return 0
}

func (k tableKey) String() string {
	if k.typ == INDEX {
		return strconv.Itoa(k.index)
	}
	return k.key
}

func (k tableKey) Describe() string {
	if k.typ == INDEX {
		return "INDEX:" + strconv.Itoa(k.index)
	}
	return "KEY:" + k.key
}

type tableEntry struct {
	key     tableKey
	rev     int
	op    	opCode
	value 	Value
}

type tableValue struct {
	typ      	TableType
	entries  	[]*tableEntry
	revision 	int         // the same as the last operation revision
	maxIndex 	int
	sorted   	bool
}

func (t *tableValue) Equal(val Value) bool {
	if val == nil || val.Kind() != TABLE {
		return false
	}
	o := val.(*tableValue)
	return reflect.DeepEqual(t.Map(), o.Map())
}

func (t tableValue) Len() int {
	return len(t.entries)
}

func (t *tableValue) Swap(i, j int) {
	t.entries[i], t.entries[j] = t.entries[j], t.entries[i]
}

func (t tableValue) Less(i, j int) bool {

	l := t.entries[i]
	r := t.entries[j]

	c := l.key.Compare(r.key)

	if c == 0 {
		c = r.rev - l.rev
	}

	if c < 0 {
		return true
	}

	return false
}

func List() *tableValue {
	return newTable(LIST)
}

func Map() *tableValue {
	return newTable(MAP)
}

func newTable(typ TableType) *tableValue {
	return &tableValue{typ: typ, entries: make([]*tableEntry, 0, InitTableSize), sorted: true}
}

func (t tableValue) Kind() Kind {
	return TABLE
}

func (t tableValue) Class() reflect.Type {
	return reflect.TypeOf((*tableValue)(nil)).Elem()
}

func (t *tableValue) String() string {
	t.sortIfNeeded()
	return "{}"
}

func (t *tableValue) Pack(p Packer) {
	t.sortIfNeeded()
	p.PackNil()
}

func (t *tableValue) Json() string {
	t.sortIfNeeded()
	return "{}"
}

func (t tableValue) Type() TableType {
	return t.typ
}

func (t *tableValue) Get(key string) Value {

	if isNumericString(key) {
		index, err := strconv.Atoi(key)
		if err == nil {
			return t.GetAt(index)
		}
	}

	t.sortIfNeeded()
	n := len(t.entries)
	i := sort.Search(n, func(i int) bool {
		e := t.entries[i]
		switch e.key.typ {
		case INDEX:
			// all indexes are in front
			return false
		case KEY:
			return e.key.key >= key
		}
		return false
	});
	if i < n {
		e :=  t.entries[i]
		if e.key.key == key && e.op == PUT {
			return e.value
		}
	}
	return nil
}

func (t *tableValue) GetTable(key string) Table {
	value := t.Get(key)
	if value != nil && value.Kind() == TABLE {
		return value.(Table)
	}
	return nil
}

func (t *tableValue) GetBool(key string) Bool {
	value := t.Get(key)
	if value != nil {
		if value.Kind() == BOOL {
			return value.(Bool)
		}
		return ParseBoolean(value.String())
	}
	return nil
}

func (t *tableValue) GetNumber(key string) Number {
	value := t.Get(key)
	if value != nil {
		if value.Kind() == NUMBER {
			return value.(Number)
		}
		return ParseNumber(value.String())
	}
	return nil
}

func (t *tableValue) GetString(key string) String {
	value := t.Get(key)
	if value != nil {
		if value.Kind() == STRING {
			return value.(String)
		}
		return ParseString(value.String())
	}
	return nil
}

func (t *tableValue) GetAt(index int) Value {
	t.sortIfNeeded()
	n := len(t.entries)
	i := sort.Search(n, func(i int) bool {
		e := t.entries[i]
		switch e.key.typ {
			case INDEX:
				return e.key.index >= index
			case KEY:
				// all keys are in the back
				return true
		}
		return false
	});
	if i < n {
		e :=  t.entries[i]
		if e.key.index == index && e.op == PUT {
			return e.value
		}
	}
	return nil
}

func (t *tableValue) GetTableAt(index int) Table {
	value := t.GetAt(index)
	if value != nil && value.Kind() == TABLE {
		return value.(Table)
	}
	return nil
}

func (t *tableValue) GetBoolAt(index int) Bool {
	value := t.GetAt(index)
	if value != nil {
		if value.Kind() == BOOL {
			return value.(Bool)
		}
		return ParseBoolean(value.String())
	}
	return nil
}

func (t *tableValue) GetNumberAt(index int) Number {
	value := t.GetAt(index)
	if value != nil {
		if value.Kind() == NUMBER {
			return value.(Number)
		}
		return ParseNumber(value.String())
	}
	return nil
}

func (t *tableValue) GetStringAt(index int) String {
	value := t.GetAt(index)
	if value != nil {
		if value.Kind() == STRING {
			return value.(String)
		}
		return ParseString(value.String())
	}
	return nil
}

func (t *tableValue) GetExp(e Expr) Value {
	return t.evaluate(e, false, func(table *tableValue, key string) Value {
		if table != nil {
			return table.Get(key)
		}
		return nil
	})
}

func (t *tableValue) GetTableExp(e Expr) Table {
	value := t.GetExp(e)
	if value != nil && value.Kind() == TABLE {
		return value.(Table)
	}
	return nil
}

func (t *tableValue) GetBoolExp(e Expr) Bool {
	value := t.GetExp(e)
	if value != nil {
		if value.Kind() == BOOL {
			return value.(Bool)
		}
		return ParseBoolean(value.String())
	}
	return nil
}

func (t *tableValue) GetNumberExp(e Expr) Number {
	value := t.GetExp(e)
	if value != nil {
		if value.Kind() == NUMBER {
			return value.(Number)
		}
		return ParseNumber(value.String())
	}
	return nil
}

func (t *tableValue) GetStringExp(e Expr) String {
	value := t.GetExp(e)
	if value != nil {
		if value.Kind() == STRING {
			return value.(String)
		}
		return ParseString(value.String())
	}
	return nil
}

func (t *tableValue) Insert(value Value) {
	t.PutAt(t.MaxIndex()+1, value)
}

func (t *tableValue) Put(key string, value Value) {

	if isNumericString(key) {
		index, err := strconv.Atoi(key)
		if err == nil {
			t.PutAt(index, value)
			return
		}
	}

	t.typ = MAP

	entry := &tableEntry{ key: tableKey{typ: KEY, key: key}, rev: t.nextRevision(), op: PUT, value: value}
	if value == nil {
		entry.op = REMOVE
	}

	t.entries = append(t.entries, entry)

	t.sorted = false
}

func (t *tableValue) PutAt(index int, value Value) {

	if t.maxIndex < index {
		t.maxIndex = index
	}

	entry := &tableEntry{ key: tableKey{typ: INDEX, index: index}, rev: t.nextRevision(), op: PUT, value: value}
	if value == nil {
		entry.op = REMOVE
	}

	t.entries = append(t.entries, entry)

	t.sorted = false
}

func (t *tableValue) PutExp(exp Expr, value Value) {

	t.evaluate(exp, true, func(table *tableValue, key string) Value {
		table.Put(key, value)
		return nil
	})

}

func (t *tableValue) Remove(key string) {

	if isNumericString(key) {
		index, err := strconv.Atoi(key)
		if err == nil {
			t.RemoveAt(index)
			return
		}
	}

	entry := &tableEntry{ key: tableKey{typ: KEY, key: key}, rev: t.nextRevision(), op: REMOVE}
	t.entries = append(t.entries, entry)

	t.sorted = false
}

func (t *tableValue) RemoveAt(index int) {

	entry := &tableEntry{ key: tableKey{typ: INDEX, index: index}, rev: t.nextRevision(), op: REMOVE}
	t.entries = append(t.entries, entry)

	t.sorted = false
}

func (t *tableValue) RemoveExp(exp Expr) {

	t.evaluate(exp, false, func(table *tableValue, key string) Value {
		if table != nil {
			table.Remove(key)
		}
		return nil
	})

}

func (t *tableValue) Map() map[string]Value {
	m := make(map[string]Value)
	t.entryProcessor(func (e *tableEntry) {
		m[e.key.String()] = e.value
	})
	return m
}

func (t *tableValue) List() []Value {
	list := make([]Value, 0, len(t.entries))
	t.entryProcessor(func (e *tableEntry) {
		list = append(list, e.value)
	})
	return list
}

func (t *tableValue) Keys() []string {
	list := make([]string, 0, len(t.entries))
	t.entryProcessor(func (e *tableEntry) {
		list = append(list, e.key.String())
	})
	return list
}

func (t *tableValue) Indexes() []int {
	list := make([]int, 0, len(t.entries))
	t.entryProcessor(func (e *tableEntry) {
		if e.key.typ == INDEX {
			list = append(list, e.key.index)
		}
	})
	return list
}

func (t tableValue) MaxIndex() int {
	return t.maxIndex
}

func (t *tableValue) Size() int {
	size := 0
	t.entryProcessor(func (e *tableEntry) {
		size = size + 1
	})
	return size
}

func (t *tableValue) Clear() {
	t.typ = LIST
	t.entries = make([]*tableEntry, 0, InitTableSize)
	t.revision = 0
	t.maxIndex = 0
	t.sorted = true
}

func (t tableValue) Sorted() bool {
	return t.sorted
}

func (t *tableValue) sortIfNeeded() {
	if !t.sorted {
		sort.Sort(t)
		t.sorted = true
	}
}

func (t *tableValue) nextRevision() int {
	t.revision = t.revision + 1
	return t.revision
}

type entryCallback = func (*tableEntry)

func (t *tableValue) entryProcessor(cb entryCallback) {
	t.sortIfNeeded()
	var k *tableKey
	for _, e := range t.entries {
		if k != nil && e.key.Compare(*k) == 0 {
			// high revision always on top
			continue
		}
		k = &e.key
		if e.op == PUT {
			cb(e)
		}
	}
}

type operationFunc = func (table *tableValue, key string) Value

func (t *tableValue) evaluate(ve Expr, createSubTables bool, op operationFunc) Value {

	if ve.Empty() {
		return nil
	}

	lastIndex := ve.Size() - 1

	current := t;
	for i := 0; i < lastIndex; i++ {

		key := ve.GetAt(i)
		val := current.Get(key)

		if val == nil || val.Kind() != TABLE {

			if createSubTables {
				newTable := List()
				current.Put(key, newTable)
				current = newTable
			} else {
				op(nil, key)
				return nil
			}

		} else {
			current = val.(*tableValue)
		}

	}

	key := ve.GetAt(lastIndex)
	return op(current, key)

}

func (t tableValue) Describe() string {
	var b strings.Builder
	fmt.Fprintf(&b, "table: %s, revision=%d, maxIndex=%d, sorted=%t {\n", t.typ, t.revision, t.maxIndex, t.sorted)
	for i, e := range t.entries {
		fmt.Fprintf(&b,"    entry[%d]=", i)
		e.DescribeTo(&b)
		fmt.Fprint(&b, "\n")
	}
	fmt.Fprintf(&b, "}\n")
	return b.String()
}

func (e tableEntry) DescribeTo(b *strings.Builder) {
	fmt.Fprintf(b, "rev=%d, op=%s, key=%v, value=%v", e.rev, e.op, e.key.Describe(), e.value)
}

func (o opCode) String() string {
	switch o {
	case PUT:
		return "PUT"
	case REMOVE:
		return "REMOVE"
	}
	return "InvalidOp"
}

func (t TableType) String() string {
	switch t {
	case LIST:
		return "LIST"
	case MAP:
		return "MAP"
	}
	return "InvalidTable"
}

func isNumericString(str string) bool {
	for _, ch := range str {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}