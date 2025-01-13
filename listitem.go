/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package value

/**

Mutable List Item

 */

type mutableListItem struct {
	key    int
	value  Value
}

func MutableItem(key int, value Value) ListItem {
	return &mutableListItem{key, value}
}

func (t *mutableListItem) Key() int {
	return t.key
}

func (t *mutableListItem) Value() Value {
	return t.value
}

func (t *mutableListItem) Update(updater Updater) bool {
	newValue := updater.Update(t.value)
	if newValue == nil {
		newValue = Null
	}
	t.value = newValue
	return true
}

func (t *mutableListItem) Equal(e ListItem) bool {
	return t.key == e.Key() && Equal(t.value, e.Value())
}


/**

Immutable List Item

*/

type immutableListItem struct {
	key    int
	value  Value
}

func ImmutableItem(key int, value Value) ListItem {
	return &immutableListItem{key, value}
}

func (t immutableListItem) Key() int {
	return t.key
}

func (t immutableListItem) Value() Value {
	return t.value
}

func (t immutableListItem) Update(Updater) bool {
	return false
}

func (t immutableListItem) Equal(e ListItem) bool {
	return t.key == e.Key() && Equal(t.value, e.Value())
}

/**

Sortable List Items

*/

type sortableItems []ListItem

func (t sortableItems) Len() int {
	return len(t)
}

func (t sortableItems) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t sortableItems) Less(i, j int) bool {
	return t[i].Key() < t[j].Key()
}