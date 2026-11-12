/*
 * Copyright (c) 2026 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value_test

import (
	"bytes"
	"encoding/hex"
	"math"
	"math/big"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	val "go.arpabet.com/value"
)

type unsignedFields struct {
	U   uint              `value:"u"`
	U64 uint64            `value:"u64"`
	P   *uint64           `value:"p,omitempty"`
	L   []uint64          `value:"l,omitempty"`
	M   map[string]uint64 `value:"m,omitempty"`
}

func bigUint(u uint64) *big.Int { return new(big.Int).SetUint64(u) }

// An unsigned field holds the same number before and after it is written,
// whatever its size: a long while it fits one, a big integer above.
func TestUnsignedFieldKeepsItsNumber(t *testing.T) {
	for _, u := range []uint64{0, 1, math.MaxInt64, math.MaxInt64 + 1, math.MaxUint64 - 1, math.MaxUint64} {
		in := unsignedFields{U: uint(u), U64: u, P: &u, L: []uint64{u, 7}, M: map[string]uint64{"k": u}}
		v, err := val.Marshal(in)
		require.NoError(t, err)

		m := v.(val.Map)
		n := m.GetNumber("u64")
		got, err := val.ExactBigInt(n)
		require.NoError(t, err)
		require.Zero(t, got.Cmp(bigUint(u)), "the value holds %s for the field %d", n, u)
		if u <= math.MaxInt64 {
			require.Equal(t, val.LONG, n.Type(), "%d", u)
		} else {
			require.Equal(t, val.BIGINT, n.Type(), "%d", u)
		}

		packed, err := val.Pack(v)
		require.NoError(t, err)
		back, err := val.Unpack(packed, true)
		require.NoError(t, err)

		for name, codec := range map[string]val.ValueCodec{
			"default":  val.DefaultValueCodec(),
			"lossless": val.DefaultValueCodec().Lossless(),
		} {
			var out unsignedFields
			require.NoError(t, codec.Unmarshal(back, &out), "%s codec, %d", name, u)
			require.Equal(t, in.U, out.U, name)
			require.Equal(t, in.U64, out.U64, name)
			require.NotNil(t, out.P, name)
			require.Equal(t, u, *out.P, name)
			require.Equal(t, in.L, out.L, name)
			require.Equal(t, in.M, out.M, name)
		}
	}
}

// The bytes of an unsigned field above the range of a long are those of the
// big integer it is, and a field within the range keeps the bytes it had.
func TestUnsignedFieldBytes(t *testing.T) {
	type one struct {
		N uint64 `value:"n"`
	}
	cases := map[uint64]string{
		0:                 "81a16e00",
		math.MaxInt64:     "81a16ecf7fffffffffffffff",
		math.MaxInt64 + 1: "81a16ec70901008000000000000000",
		math.MaxUint64:    "81a16ec7090100ffffffffffffffff",
	}
	for u, want := range cases {
		v, err := val.Marshal(one{N: u})
		require.NoError(t, err)
		packed, err := val.Pack(v)
		require.NoError(t, err)
		require.Equal(t, want, hex.EncodeToString(packed), "%d", u)
	}
}

// Unsigned makes the number of an unsigned integer.
func TestUnsignedConstructor(t *testing.T) {
	require.True(t, val.Unsigned(5).Equal(val.Long(5)))
	require.True(t, val.Unsigned(math.MaxInt64).Equal(val.Long(math.MaxInt64)))
	require.True(t, val.Unsigned(math.MaxUint64).Equal(val.BigInt(bigUint(math.MaxUint64))))

	u, err := val.ExactUint(val.Unsigned(math.MaxUint64), 64)
	require.NoError(t, err)
	require.Equal(t, uint64(math.MaxUint64), u)
}

// MessagePack's uint 64 holds integers a long does not. A canonical writer
// never uses it for them, another MessagePack writer does: they read as the
// number they are and pack again as that big integer.
func TestUint64FormAboveLongReadsAsItsNumber(t *testing.T) {
	cases := []struct {
		hex    string
		want   uint64
		typ    val.NumberType
		repack string
	}{
		{"cf7fffffffffffffff", math.MaxInt64, val.LONG, "cf7fffffffffffffff"},
		{"cf8000000000000000", math.MaxInt64 + 1, val.BIGINT, "c70901008000000000000000"},
		{"cfffffffffffffffff", math.MaxUint64, val.BIGINT, "c7090100ffffffffffffffff"},
	}
	for _, c := range cases {
		raw, err := hex.DecodeString(c.hex)
		require.NoError(t, err)

		v, err := val.Unpack(raw, true)
		require.NoError(t, err, c.hex)
		n, ok := v.(val.Number)
		require.True(t, ok, c.hex)
		require.Equal(t, c.typ, n.Type(), c.hex)
		got, err := val.ExactBigInt(n)
		require.NoError(t, err)
		require.Zero(t, got.Cmp(bigUint(c.want)), "%s reads as %s", c.hex, n)

		packed, err := val.Pack(v)
		require.NoError(t, err)
		require.Equal(t, c.repack, hex.EncodeToString(packed), c.hex)

		// The reader of a stream gives the same value.
		streamed, err := val.Read(bytes.NewReader(raw))
		require.NoError(t, err, c.hex)
		require.True(t, v.Equal(streamed), "%s read from a stream as %s", c.hex, streamed)
	}

	// Inside a struct field, with either codec.
	raw, err := hex.DecodeString("81a16ecfffffffffffffffff")
	require.NoError(t, err)
	v, err := val.Unpack(raw, true)
	require.NoError(t, err)
	for _, codec := range []val.ValueCodec{val.DefaultValueCodec(), val.DefaultValueCodec().Lossless()} {
		var out struct {
			N uint64 `value:"n"`
		}
		require.NoError(t, codec.Unmarshal(v, &out))
		require.Equal(t, uint64(math.MaxUint64), out.N)
	}
}

// Data written before v1.6.1 holds an unsigned field above the range of a long
// as the negative long of the same bits. The default codec still reads it, as
// it wraps every number; the lossless codec refuses a negative number for an
// unsigned field, as CANONICAL.md says.
func TestUnsignedFieldWrittenBefore(t *testing.T) {
	type one struct {
		N uint64 `value:"n"`
	}
	old := val.ImmutableMapOf(map[string]val.Value{"n": val.Long(-1)})

	var out one
	require.NoError(t, val.DefaultValueCodec().Unmarshal(old, &out))
	require.Equal(t, uint64(math.MaxUint64), out.N)

	require.Error(t, val.DefaultValueCodec().Lossless().Unmarshal(old, &out))
}

// The default codec wraps a big integer beyond the field to its low bits, as
// it does a long.
func TestDefaultCodecWrapsBigIntegers(t *testing.T) {
	type one struct {
		N uint64 `value:"n"`
	}
	cases := map[string]uint64{
		"18446744073709551615": math.MaxUint64, // 2^64 - 1 fits
		"18446744073709551616": 0,              // 2^64 wraps to 0
		"18446744073709551621": 5,              // 2^64 + 5
		"-1":                   math.MaxUint64, // as a long -1 does
	}
	for text, want := range cases {
		x, ok := new(big.Int).SetString(text, 10)
		require.True(t, ok)
		var out one
		require.NoError(t, val.Unmarshal(val.ImmutableMapOf(map[string]val.Value{"n": val.BigInt(x)}), &out), text)
		require.Equal(t, want, out.N, text)
	}
	// A decimal goes by its integer part, as before.
	var out one
	require.NoError(t, val.Unmarshal(val.ImmutableMapOf(map[string]val.Value{"n": val.Decimal(decimal.RequireFromString("18446744073709551615"))}), &out))
	require.Equal(t, uint64(math.MaxUint64), out.N)
	require.NoError(t, val.Unmarshal(val.ImmutableMapOf(map[string]val.Value{"n": val.Decimal(decimal.RequireFromString("7.9"))}), &out))
	require.Equal(t, uint64(7), out.N)
}

// An integer text beyond the range of a long is the big integer it spells,
// not the nearest double.
func TestParseNumberBeyondLong(t *testing.T) {
	for _, text := range []string{"9223372036854775808", "-9223372036854775809", "18446744073709551615", "+18446744073709551615", "123456789012345678901234567890"} {
		n := val.ParseNumber(text)
		require.Equal(t, val.BIGINT, n.Type(), text)
		want, ok := new(big.Int).SetString(text, 10)
		require.True(t, ok)
		require.Zero(t, n.BigInt().Cmp(want), "%s parsed as %s", text, n)
	}
	// What a long holds stays a long, and other texts keep their reading.
	require.True(t, val.ParseNumber("9223372036854775807").Equal(val.Long(math.MaxInt64)))
	require.True(t, val.ParseNumber("-9223372036854775808").Equal(val.Long(math.MinInt64)))
	require.Equal(t, val.DOUBLE, val.ParseNumber("1e30").Type())
	require.True(t, val.ParseNumber("12ab").IsNaN())
}

// BigInt of a decimal is its integer part: the whole number when the exponent
// is above zero, and rounded toward zero as Long rounds.
func TestDecimalBigInt(t *testing.T) {
	cases := []struct {
		d    decimal.Decimal
		want int64
	}{
		{decimal.New(1, 3), 1000},
		{decimal.New(-25, 2), -2500},
		{decimal.RequireFromString("123.99"), 123},
		{decimal.RequireFromString("-1.5"), -1},
		{decimal.RequireFromString("-0.5"), 0},
		{decimal.RequireFromString("42"), 42},
	}
	for _, c := range cases {
		n := val.Decimal(c.d)
		require.Zero(t, n.BigInt().Cmp(big.NewInt(c.want)), "BigInt of %s is %s", c.d, n.BigInt())
		require.Equal(t, c.want, n.Long(), "Long of %s", c.d)
	}
}
