/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value

func EmptyList(immutable bool) List {
	if immutable {
		return immutableListValue([]Value{})
	} else {
		return solidListValue([]Value{})
	}
}

func EmptyMap(immutable bool) Map {
	if immutable {
		return immutableMapValue([]MapEntry{})
	} else {
		return sortedMapValue([]MapEntry{})
	}
}