/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value

/**

Mutable Map Entry

*/

type mutableMapEntry struct {
	key    string
	value  Value
}

func MutableEntry(key string, value Value) MapEntry {
	return &mutableMapEntry{
		key: key,
		value: value,
	}
}

func (t *mutableMapEntry) Key() string {
	return t.key
}

func (t *mutableMapEntry) Value() Value {
	return t.value
}

func (t *mutableMapEntry) Update(updater Updater) bool {
	newValue := updater.Update(t.value)
	if newValue == nil {
		newValue = Null
	}
	t.value = newValue
	return true
}

func (t *mutableMapEntry) Equal(e MapEntry) bool {
	return t.key == e.Key() && Equal(t.value, e.Value())
}

/**

Immutable Map Entry

*/

type immutableMapEntry struct {
	key    string
	value  Value
}

func ImmutableEntry(key string, value Value) MapEntry {
	return &immutableMapEntry {
		key: key,
		value: value,
	}
}


func (t *immutableMapEntry) Key() string {
	return t.key
}

func (t *immutableMapEntry) Value() Value {
	return t.value
}

func (t *immutableMapEntry) Update(Updater) bool {
	return false
}

func (t *immutableMapEntry) Equal(e MapEntry) bool {
	return t.key == e.Key() && Equal(t.value, e.Value())
}

func mapEntryCopyOf(src []MapEntry) []MapEntry {
	n := len(src)
	dst := make([]MapEntry, n)
	copy(dst, src)
	return dst
}