/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value_test

import (
	"math/big"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	val "go.arpabet.com/value"
)

// TestCanonicalGoldenVectors locks the canonical MessagePack encoding of
// representative values. A change to any byte here is a change to the hashable
// form and must be deliberate. See CANONICAL.md.
func TestCanonicalGoldenVectors(t *testing.T) {
	cases := []struct {
		name string
		v    val.Value
		hex  string
	}{
		{"null", val.Null, "c0"},
		{"true", val.Boolean(true), "c3"},
		{"false", val.Boolean(false), "c2"},
		{"long-0", val.Long(0), "00"},
		{"long-127", val.Long(127), "7f"},
		{"long-128", val.Long(128), "cc80"},
		{"long-neg1", val.Long(-1), "ff"},
		{"long-neg33", val.Long(-33), "d0df"},
		{"double-0", val.Double(0), "cb0000000000000000"},
		{"double-1.5", val.Double(1.5), "cb3ff8000000000000"},
		{"bigint-0", val.BigInt(big.NewInt(0)), "d40102"},
		{"bigint-123", val.BigInt(big.NewInt(123)), "d501027b"},
		{"utf8-empty", val.Utf8(""), "a0"},
		{"utf8-hello", val.Utf8("hello"), "a568656c6c6f"},
		{"raw-123", val.Raw([]byte{1, 2, 3}, false), "c403010203"},
		{"list-empty", val.EmptyImmutableList(), "90"},
		{"list-123", val.ImmutableList([]val.Value{val.Long(1), val.Long(2), val.Long(3)}), "93010203"},
		{"map-empty", val.EmptyImmutableMap(), "80"},
		{"map-ab", val.ImmutableMapOf(map[string]val.Value{"a": val.Long(1), "b": val.Long(2)}), "82a16101a16202"},
		{"map-nested", val.ImmutableMapOf(map[string]val.Value{"list": val.ImmutableList([]val.Value{val.Long(1), val.Long(2)})}), "81a46c697374920102"},
	}
	for _, c := range cases {
		require.Equal(t, c.hex, val.Hex(c.v), "canonical bytes changed for %q", c.name)
	}
}

// TestCanonicalMapOrderIndependent verifies that map encoding depends only on
// the contents, not on insertion order: keys are always sorted.
func TestCanonicalMapOrderIndependent(t *testing.T) {
	m1 := val.EmptyImmutableMap().Put("b", val.Long(2)).Put("a", val.Long(1))
	m2 := val.EmptyImmutableMap().Put("a", val.Long(1)).Put("b", val.Long(2))
	require.Equal(t, val.Hex(m2), val.Hex(m1))
	require.Equal(t, "82a16101a16202", val.Hex(m1))
}

// TestCanonicalIdempotent verifies the canonical form is a fixed point of
// Pack∘Unpack: decoding then re-encoding reproduces identical bytes.
func TestCanonicalIdempotent(t *testing.T) {
	values := []val.Value{
		val.Null,
		val.Boolean(true),
		val.Long(123456789),
		val.Double(-12.34),
		val.BigInt(big.NewInt(-1000)),
		val.Decimal(decimal.New(12345, -2)),
		val.Utf8("hello"),
		val.Raw([]byte{0, 1, 2}, false),
		val.ImmutableList([]val.Value{val.Long(1), val.Utf8("x")}),
		val.ImmutableMapOf(map[string]val.Value{"a": val.Long(1), "b": val.Boolean(false)}),
	}
	for _, v := range values {
		first, err := val.Pack(v)
		require.NoError(t, err)
		decoded, err := val.Unpack(first, false)
		require.NoError(t, err)
		second, err := val.Pack(decoded)
		require.NoError(t, err)
		require.Equal(t, first, second, "Pack/Unpack not idempotent for %s", v.String())
	}
}

// TestCanonicalEqualImpliesEqualBytes checks the content-addressing invariant
// across the value tree: equal values of the same kind serialize identically,
// regardless of their concrete (mutable vs immutable) representation.
func TestCanonicalEqualImpliesEqualBytes(t *testing.T) {
	pairs := [][2]val.Value{
		{val.Boolean(true), val.Boolean(true)},
		{val.Long(5), val.Long(5)},
		{val.Utf8("x"), val.Utf8("x")},
		{val.MutableList([]val.Value{val.Long(1)}), val.ImmutableList([]val.Value{val.Long(1)})},
		{val.SortedMapOf(map[string]val.Value{"a": val.Long(1)}), val.ImmutableMapOf(map[string]val.Value{"a": val.Long(1)})},
	}
	for _, p := range pairs {
		require.True(t, p[0].Equal(p[1]), "expected equal: %s", p[0].String())
		a, err := val.Pack(p[0])
		require.NoError(t, err)
		b, err := val.Pack(p[1])
		require.NoError(t, err)
		require.Equal(t, a, b, "equal values must pack identically: %s", p[0].String())
	}
}
