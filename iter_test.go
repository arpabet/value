/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	val "go.arpabet.com/value"
)

func TestAsGeneric(t *testing.T) {
	v := val.Value(val.Long(7))

	n, ok := val.As[val.Number](v)
	require.True(t, ok)
	require.Equal(t, int64(7), n.Long())

	_, ok = val.As[val.Map](v)
	require.False(t, ok)
}

func TestMapAll(t *testing.T) {
	m := val.EmptyImmutableMap().Put("b", val.Long(2)).Put("a", val.Long(1))

	var keys []string
	got := map[string]int64{}
	for k, v := range val.MapAll(m) {
		keys = append(keys, k)
		got[k] = v.(val.Number).Long()
	}

	require.Equal(t, []string{"a", "b"}, keys) // canonical sorted order
	require.Equal(t, int64(1), got["a"])
	require.Equal(t, int64(2), got["b"])
}

func TestListAll(t *testing.T) {
	l := val.ImmutableList([]val.Value{val.Long(10), val.Long(20), val.Long(30)})

	var sum int64
	count := 0
	for i, v := range val.ListAll(l) {
		sum += int64(i) * v.(val.Number).Long()
		count++
	}

	require.Equal(t, 3, count)
	require.Equal(t, int64(0*10+1*20+2*30), sum)
}

func TestListAllEarlyBreak(t *testing.T) {
	l := val.ImmutableList([]val.Value{val.Long(1), val.Long(2), val.Long(3)})

	count := 0
	for range val.ListAll(l) {
		count++
		break
	}
	require.Equal(t, 1, count) // break must stop iteration
}
