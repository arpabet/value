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

// A tiny header claiming a huge collection must be rejected, not pre-allocated.

func TestUnpackRejectsHugeArrayHeader(t *testing.T) {
	blob := []byte{0xdd, 0xff, 0xff, 0xff, 0xff} // array32, ~4.29e9 elements, no payload
	_, err := val.Unpack(blob, false)
	require.Error(t, err)
}

func TestUnpackRejectsHugeMapHeader(t *testing.T) {
	blob := []byte{0xdf, 0xff, 0xff, 0xff, 0xff} // map32, ~4.29e9 entries, no payload
	_, err := val.Unpack(blob, false)
	require.Error(t, err)
}

func TestUnpackRejectsHugeStringHeader(t *testing.T) {
	blob := []byte{0xdb, 0xff, 0xff, 0xff, 0xff} // str32, ~4.29e9 bytes, no payload
	_, err := val.Unpack(blob, false)
	require.Error(t, err)
}

// Deeply nested input must be rejected before exhausting the stack.

func TestUnpackRejectsDeepNesting(t *testing.T) {
	n := val.MaxParseDepth + 10
	blob := make([]byte, n)
	for i := range blob {
		blob[i] = 0x91 // fixarray holding one element -> the next array
	}
	_, err := val.Unpack(blob, false)
	require.Error(t, err)
}

// The limits are configurable and actually enforced.

func TestParseCollectionLimitIsConfigurable(t *testing.T) {
	defer func(old int) { val.MaxParseCollectionLen = old }(val.MaxParseCollectionLen)
	val.MaxParseCollectionLen = 2

	_, err := val.Unpack([]byte{0x93, 0x01, 0x02, 0x03}, false) // fixarray of 3
	require.Error(t, err)

	v, err := val.Unpack([]byte{0x92, 0x01, 0x02}, false) // fixarray of 2 (within limit)
	require.NoError(t, err)
	require.Equal(t, 2, v.(val.List).Len())
}

func TestUnpackHugeSparseIndexBecomesMap(t *testing.T) {
	// A map with one enormous integer key must NOT become a sparse list, whose
	// Len()/Values() would otherwise try to allocate ~maxKey entries.
	blob := []byte{0x81, 0xcf, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}
	out, err := val.Unpack(blob, false)
	require.NoError(t, err)
	require.Equal(t, val.MAP, out.Kind())
	m := out.(val.Map)
	require.Equal(t, 1, m.Len())
	require.Equal(t, int64(1), m.GetNumber("9223372036854775807").Long())
}

func TestParseByteLimitIsConfigurable(t *testing.T) {
	defer func(old int) { val.MaxParseByteLen = old }(val.MaxParseByteLen)
	val.MaxParseByteLen = 4

	_, err := val.Unpack([]byte{0xd9, 0x05, 'h', 'e', 'l', 'l', 'o'}, false) // str8 len 5
	require.Error(t, err)
}
