/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value_test

import (
	"math"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	val "go.arpabet.com/value"
)

func dec(t *testing.T, s string) val.Number {
	t.Helper()
	d, err := decimal.NewFromString(s)
	require.NoError(t, err)
	return val.Decimal(d)
}

// packed returns the canonical bytes, which tell kinds and exact digits apart
// where Equal compares numbers by value.
func packed(t *testing.T, v val.Value) string {
	t.Helper()
	b, err := val.Pack(v)
	require.NoError(t, err)
	return string(b)
}

func everyKind(t *testing.T) val.Value {
	big1, _ := new(big.Int).SetString("-123456789012345678901234567890", 10)
	return val.ImmutableMapOf(map[string]val.Value{
		"bool":    val.Boolean(true),
		"null":    val.Null,
		"long":    val.Long(-42),
		"double":  val.Double(1.5),
		"tiny":    val.Double(1e-300),
		"bigint":  val.BigInt(big1),
		"decimal": dec(t, "1.10"),
		"huge":    dec(t, "1e400"),
		"str":     val.Utf8("quote \" backslash \\ newline \n tab \t nul \x00 snow ☃ emoji 😀"),
		"bin":     val.Raw([]byte{0, 1, 2, 0xfe, 0xff}, false),
		"time":    val.Time(time.Date(2026, 10, 30, 7, 28, 39, 75000001, time.UTC)),
		"list":    val.ImmutableList([]val.Value{val.Long(1), val.Utf8("two"), val.EmptyImmutableMap()}),
		"sparse":  val.SparseList([]val.ListItem{val.ImmutableItem(0, val.Utf8("a")), val.ImmutableItem(5, val.Utf8("b"))}, true),
		"ext":     val.Unknown([]byte{7, 1, 2, 3}),
		"escaped": val.ImmutableMapOf(map[string]val.Value{"$bin": val.Utf8("not binary")}),
		"nested":  val.ImmutableMapOf(map[string]val.Value{"$map": val.ImmutableMapOf(map[string]val.Value{"$long": val.Utf8("5")})}),
		"empty":   val.EmptyImmutableList(),
	})
}

func TestJSONCanonicalRoundTrip(t *testing.T) {
	v := everyKind(t)
	text, err := val.FormatJSON(v, val.JSONCanonical)
	require.NoError(t, err)
	back, err := val.ParseJSON(text)
	require.NoError(t, err, string(text))
	require.Equal(t, packed(t, v), packed(t, back), string(text))
	again, err := val.FormatJSON(back, val.JSONCanonical)
	require.NoError(t, err)
	require.Equal(t, string(text), string(again))
}

func TestJSONCanonicalForms(t *testing.T) {
	cases := []struct {
		v    val.Value
		want string
	}{
		{val.Long(5), `{"$long":"5"}`},
		{val.Double(1.5), `{"$double":"1.5"}`},
		{val.Double(math.Copysign(0, -1)), `{"$double":"-0"}`},
		{val.Double(math.Inf(-1)), `{"$double":"-Infinity"}`},
		{val.BigInt(big.NewInt(-7)), `{"$bigint":"-7"}`},
		{dec(t, "1.10"), `{"$decimal":"1.10"}`},
		{val.Raw([]byte{1, 2}, false), `{"$bin":"AQI="}`},
		{val.Time(time.Date(2026, 10, 30, 7, 28, 39, 75000000, time.UTC)), `{"$timestamp":"2026-10-30T07:28:39.075Z"}`},
		{val.Unknown([]byte{0xf9, 1}), `{"$ext":{"base64":"AQ==","type":-7}}`},
		{val.SparseList([]val.ListItem{val.ImmutableItem(10, val.Boolean(true)), val.ImmutableItem(2, val.Null)}, false), `{"$sparse":{"2":null,"10":true}}`},
		{val.ImmutableMapOf(map[string]val.Value{"$map": val.Long(1)}), `{"$map":{"$map":{"$long":"1"}}}`},
		{val.ImmutableMapOf(map[string]val.Value{"b": val.Utf8("x"), "a": val.Null}), `{"a":null,"b":"x"}`},
	}
	for _, c := range cases {
		got, err := val.FormatJSON(c.v, val.JSONCanonical)
		require.NoError(t, err)
		require.Equal(t, c.want, string(got))
		back, err := val.ParseJSON(got)
		require.NoError(t, err, c.want)
		require.Equal(t, packed(t, c.v), packed(t, back), c.want)
	}
}

func TestJSONRelaxedNumbers(t *testing.T) {
	cases := []struct {
		v    val.Value
		want string
	}{
		{val.Long(-42), `-42`},
		{val.Double(1.5), `1.5`},
		{val.Double(100), `100`},
		{val.Double(1e21), `1e+21`},
		{val.Double(1e20), `100000000000000000000`},
		{val.Double(1e-7), `1e-7`},
		{val.Double(0.000001), `0.000001`},
		{val.Double(-2.5e-10), `-2.5e-10`},
		{val.Double(math.NaN()), `{"$double":"NaN"}`},
		{dec(t, "1.10"), `1.10`},
		{dec(t, "0.005"), `0.005`},
		{dec(t, "-0.00"), `0.00`},
		{dec(t, "1e3"), `1e3`},
		{dec(t, "1e-10"), `1e-10`},
		{dec(t, "1e400"), `1e400`},
		{val.BigInt(big.NewInt(1)), `{"$bigint":"1"}`},
	}
	for _, c := range cases {
		got, err := val.FormatJSON(c.v, val.JSONRelaxed)
		require.NoError(t, err)
		require.Equal(t, c.want, string(got))
	}
}

// A bare number carries no type: it reads as a Decimal with its exact digits.
func TestJSONUntypedNumbersAreDecimal(t *testing.T) {
	v, err := val.ParseJSON([]byte(` {"qty": 5, "price": 1.10, "big": 123456789012345678901234567890, "exp": -1.5E-3, "z": -0} `))
	require.NoError(t, err)
	m := v.(val.Map)
	for key, want := range map[string]string{"qty": "5", "price": "1.10", "big": "123456789012345678901234567890", "exp": "-0.0015", "z": "0"} {
		n, ok := m.Get(key).(val.Number)
		require.True(t, ok, key)
		require.Equal(t, val.DECIMAL, n.Type(), key)
		require.Equal(t, packed(t, dec(t, want)), packed(t, n), key)
	}
	relaxed, err := val.FormatJSON(v, val.JSONRelaxed)
	require.NoError(t, err)
	require.Equal(t, `{"big":123456789012345678901234567890,"exp":-0.0015,"price":1.10,"qty":5,"z":0}`, string(relaxed))
}

func TestJSONRefused(t *testing.T) {
	for _, text := range []string{
		``, `{`, `[1,]`, `{"a":1,}`, `01`, `+1`, `.5`, `1.`, `1e`, `-`, `tru`, `nul`, `"abc`,
		`{"a":1,"a":2}`, `"\ud800"`, `"\udc00x"`, `"\ud800A"`, "\"\x01\"", "\"\xff\"", `"\x"`, `1 2`,
		`{"$long":"x"}`, `{"$long":5}`, `{"$long":"1.5"}`, `{"$long":"99999999999999999999"}`,
		`{"$double":"nan"}`, `{"$double":"0x10"}`, `{"$double":"1e999"}`, `{"$bigint":"1.0"}`,
		`{"$decimal":"abc"}`, `{"$decimal":"1e99999999"}`, `{"$bin":"!!"}`, `{"$timestamp":"yesterday"}`,
		`{"$sparse":[1]}`, `{"$sparse":{"01":1}}`, `{"$sparse":{"-1":1}}`, `{"$ext":{"type":200,"base64":""}}`,
		`{"$ext":{"type":1.5,"base64":""}}`, `{"$ext":{"base64":""}}`, `{"$map":1}`, `{"$timestamp":"2026-10-30T07:28:39Z", "x":`,
	} {
		_, err := val.ParseJSON([]byte(text))
		require.Error(t, err, text)
	}
	_, err := val.FormatJSON(val.Utf8("\xff"), val.JSONRelaxed)
	require.Error(t, err)
	deep := strings.Repeat("[", val.MaxParseDepth+2) + strings.Repeat("]", val.MaxParseDepth+2)
	_, err = val.ParseJSON([]byte(deep))
	require.Error(t, err)
}

func TestJSONWrappersAmongKeys(t *testing.T) {
	// A wrapper key next to other keys is an ordinary key.
	v, err := val.ParseJSON([]byte(`{"$long":"5","other":true}`))
	require.NoError(t, err)
	require.Equal(t, val.STRING, v.(val.Map).Get("$long").Kind())
	text, err := val.FormatJSON(v, val.JSONCanonical)
	require.NoError(t, err)
	require.Equal(t, `{"$long":"5","other":true}`, string(text))

	// A known extension type written as $ext reads as its kind.
	v, err = val.ParseJSON([]byte(`{"$ext":{"type":-1,"base64":"aeRHJw=="}}`))
	require.NoError(t, err)
	require.Equal(t, val.TIMESTAMP, v.Kind())

	// Escapes and unpadded base64 are accepted.
	v, err = val.ParseJSON([]byte(`["é😀\/", {"$bin":"AQI"}]`))
	require.NoError(t, err)
	l := v.(val.List)
	require.Equal(t, "é😀/", l.GetAt(0).(val.String).Utf8())
	require.Equal(t, []byte{1, 2}, l.GetAt(1).(val.String).Raw())
}

func FuzzParseJSON(f *testing.F) {
	for _, s := range []string{`{"a":[1,2.5,"x",null,true]}`, `{"$map":{"$bin":"x"}}`, `{"$sparse":{"3":1}}`, `{"$ext":{"base64":"AQ==","type":9}}`, `{"$double":"NaN"}`} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		v, err := val.ParseJSON(data)
		if err != nil {
			return
		}
		text, err := val.FormatJSON(v, val.JSONCanonical)
		if err != nil {
			t.Fatalf("format %q: %v", data, err)
		}
		back, err := val.ParseJSON(text)
		if err != nil {
			t.Fatalf("reparse %s: %v", text, err)
		}
		again, err := val.FormatJSON(back, val.JSONCanonical)
		if err != nil || string(again) != string(text) {
			t.Fatalf("canonical text changed: %s then %s (%v)", text, again, err)
		}
	})
}
