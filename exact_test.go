/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value_test

import (
	"math"
	"math/big"
	"reflect"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	val "go.arpabet.com/value"
)

type exactTarget struct {
	I8  int8    `value:"i8,omitempty"`
	I64 int64   `value:"i64,omitempty"`
	U16 uint16  `value:"u16,omitempty"`
	F32 float32 `value:"f32,omitempty"`
	F64 float64 `value:"f64,omitempty"`
}

func decode(t *testing.T, c val.ValueCodec, field string, n val.Number) (exactTarget, error) {
	t.Helper()
	var out exactTarget
	err := c.Unmarshal(val.ImmutableMapOf(map[string]val.Value{field: n}), &out)
	return out, err
}

func TestLosslessNumbers(t *testing.T) {
	lossless := val.DefaultValueCodec().Lossless()
	ok := []struct {
		field string
		n     val.Number
		check func(exactTarget) bool
	}{
		{"i8", dec(t, "5"), func(x exactTarget) bool { return x.I8 == 5 }},
		{"i8", dec(t, "5.00"), func(x exactTarget) bool { return x.I8 == 5 }},
		{"i8", val.Long(-128), func(x exactTarget) bool { return x.I8 == -128 }},
		{"i8", val.Double(3), func(x exactTarget) bool { return x.I8 == 3 }},
		{"i64", val.BigInt(big.NewInt(math.MinInt64)), func(x exactTarget) bool { return x.I64 == math.MinInt64 }},
		{"u16", dec(t, "65535"), func(x exactTarget) bool { return x.U16 == 65535 }},
		{"f64", dec(t, "1.10"), func(x exactTarget) bool { return x.F64 == 1.1 }},
		{"f64", val.Long(1 << 53), func(x exactTarget) bool { return x.F64 == 1<<53 }},
		{"f32", val.Double(0.5), func(x exactTarget) bool { return x.F32 == 0.5 }},
		{"f32", dec(t, "0.1"), func(x exactTarget) bool { return x.F32 == float32(0.1) }},
	}
	for _, c := range ok {
		got, err := decode(t, lossless, c.field, c.n)
		require.NoError(t, err, "%s %s", c.field, c.n)
		require.True(t, c.check(got), "%s %s: %+v", c.field, c.n, got)
	}
	refused := []struct {
		field string
		n     val.Number
	}{
		{"i8", dec(t, "5.5")},
		{"i8", val.Long(128)},
		{"i8", val.Double(1.5)},
		{"i8", val.Double(math.NaN())},
		{"i64", dec(t, "9223372036854775808")},
		{"u16", val.Long(-1)},
		{"u16", val.Long(65536)},
		{"f64", dec(t, "0.12345678901234567890")},
		{"f64", val.Long(1<<53 + 1)},
		{"f64", dec(t, "1e400")},
		{"f32", val.Double(0.1)},
	}
	for _, c := range refused {
		_, err := decode(t, lossless, c.field, c.n)
		require.Error(t, err, "%s %s", c.field, c.n)
	}

	// The default codec keeps converting as before.
	got, err := decode(t, val.DefaultValueCodec(), "i8", dec(t, "5.5"))
	require.NoError(t, err)
	require.Equal(t, int8(5), got.I8)
}

type evolved struct {
	Name  string `value:"name"`
	Added int    `value:"added"`
}

func TestAllowMissing(t *testing.T) {
	old := val.ImmutableMapOf(map[string]val.Value{"name": val.Utf8("a"), "gone": val.Long(1)})
	var out evolved
	require.Error(t, val.Unmarshal(old, &out))
	require.NoError(t, val.DefaultValueCodec().AllowMissing().Unmarshal(old, &out))
	require.Equal(t, evolved{Name: "a"}, out)
}

type bigFields struct {
	Price decimal.Decimal `value:"price"`
	Count big.Int         `value:"count"`
	Opt   *big.Int        `value:"opt"`
}

func TestDecimalAndBigIntFields(t *testing.T) {
	in := bigFields{Price: decimal.RequireFromString("1.10"), Count: *big.NewInt(7)}
	in.Count.Exp(big.NewInt(10), big.NewInt(30), nil)
	v, err := val.Marshal(in)
	require.NoError(t, err)
	m := v.(val.Map)
	require.Equal(t, val.DECIMAL, m.Get("price").(val.Number).Type())
	require.Equal(t, val.BIGINT, m.Get("count").(val.Number).Type())
	require.Equal(t, val.NULL, m.Get("opt").Kind())

	b, err := val.Pack(v)
	require.NoError(t, err)
	back, err := val.Unpack(b, false)
	require.NoError(t, err)
	var out bigFields
	require.NoError(t, val.Unmarshal(back, &out))
	require.Equal(t, "1.10", val.DecimalText(out.Price))
	require.Equal(t, in.Count.String(), out.Count.String())
	require.Nil(t, out.Opt)

	// Any exact number reads into them; a fraction does not make an integer.
	src := val.ImmutableMapOf(map[string]val.Value{"price": val.Double(0.1), "count": dec(t, "12e2"), "opt": val.Long(3)})
	require.NoError(t, val.Unmarshal(src, &out))
	require.Equal(t, "0.1", out.Price.String())
	require.Equal(t, "1200", out.Count.String())
	require.Equal(t, "3", out.Opt.String())
	src = val.ImmutableMapOf(map[string]val.Value{"price": val.Long(1), "count": dec(t, "1.5"), "opt": val.Null})
	require.Error(t, val.Unmarshal(src, &out))
}

type anyField struct {
	Any any `value:"any"`
}

func TestAnyField(t *testing.T) {
	v, err := val.Marshal(anyField{Any: map[string]int{"a": 1}})
	require.NoError(t, err)
	var out anyField
	require.NoError(t, val.Unmarshal(v, &out))
	m, ok := out.Any.(val.Map)
	require.True(t, ok)
	require.Equal(t, int64(1), m.GetNumber("a").Long())
	var fresh anyField
	require.NoError(t, val.Unmarshal(val.ImmutableMapOf(map[string]val.Value{"any": val.Null}), &fresh))
	require.Nil(t, fresh.Any)
}

type tagged struct {
	A      string `value:"a"`
	B      int    `value:"b,omitempty"`
	Skip   int    `value:"-"`
	hidden int
	Plain  bool
}

func TestStructFields(t *testing.T) {
	fields, err := val.DefaultValueCodec().StructFields(reflect.TypeOf(tagged{}))
	require.NoError(t, err)
	var got []string
	for _, f := range fields {
		name := f.Name
		if f.OmitEmpty {
			name += "?"
		}
		got = append(got, name)
	}
	require.Equal(t, []string{"a", "b?", "Plain"}, got)
	_, err = val.DefaultValueCodec().StructFields(reflect.TypeOf(0))
	require.Error(t, err)
	_ = tagged{hidden: 1}
}
