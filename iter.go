/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value

import "iter"

// As returns v as the concrete value interface T (one of Bool, Number, String,
// List, Map, Extension), reporting whether v had that type. It replaces manual
// type assertions:
//
//	if n, ok := value.As[value.Number](v); ok { ... }
func As[T Value](v Value) (T, bool) {
	t, ok := v.(T)
	return t, ok
}

// MapAll returns an iterator over a Map's entries in canonical (sorted) key
// order, for use with range-over-func:
//
//	for k, v := range value.MapAll(m) { ... }
func MapAll(m Map) iter.Seq2[string, Value] {
	return func(yield func(string, Value) bool) {
		for _, e := range m.Entries() {
			if !yield(e.Key(), e.Value()) {
				return
			}
		}
	}
}

// ListAll returns an iterator over a List's values by index, for use with
// range-over-func:
//
//	for i, v := range value.ListAll(l) { ... }
func ListAll(l List) iter.Seq2[int, Value] {
	return func(yield func(int, Value) bool) {
		for i, v := range l.Values() {
			if !yield(i, v) {
				return
			}
		}
	}
}
