/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value_test

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	val "go.arpabet.com/value"
)

func bigFromString(t *testing.T, s string) *big.Int {
	x, ok := new(big.Int).SetString(s, 10)
	require.True(t, ok, s)
	return x
}

func TestBigIntPortablePayload(t *testing.T) {
	cases := []struct {
		text string
		hex  string
	}{
		{"0", "d40100"},
		{"1", "d40101"},
		{"-1", "d401ff"},
		{"127", "d4017f"},
		{"128", "d5010080"},
		{"-128", "d40180"},
		{"-129", "d501ff7f"},
		{"255", "d50100ff"},
		{"256", "d5010100"},
		{"-256", "d501ff00"},
		{"-1000", "d501fc18"},
		{"32767", "d5017fff"},
		{"32768", "c70301008000"},
		{"-32768", "d5018000"},
		{"-32769", "c70301ff7fff"},
		{"9223372036854775807", "d7017fffffffffffffff"},
		{"9223372036854775808", "c70901008000000000000000"},
		{"-9223372036854775808", "d7018000000000000000"},
		{"-9223372036854775809", "c70901ff7fffffffffffffff"},
		{"18446744073709551616", "c70901010000000000000000"},
		{"100000000000000000000", "c70901056bc75e2d63100000"},
		{"-100000000000000000000", "c70901fa9438a1d29cf00000"},
	}
	for _, c := range cases {
		v := val.BigInt(bigFromString(t, c.text))
		require.Equal(t, c.hex, val.Hex(v), c.text)
		b, err := hex.DecodeString(c.hex)
		require.NoError(t, err)
		back, err := val.Unpack(b, false)
		require.NoError(t, err, c.text)
		require.True(t, v.Equal(back), "%s: %s != %s", c.text, v, back)
		require.Equal(t, c.text, back.(val.Number).BigInt().String())
	}
	// Longer, sign-extended encodings decode to the same integer and
	// re-pack minimally.
	for _, c := range []struct{ in, out string }{
		{"d50100" + "7b", "d4017b"},
		{"c7030100007b", "d4017b"},
		{"d501ffff", "d401ff"},
		{"d501ff80", "d40180"},
	} {
		b, err := hex.DecodeString(c.in)
		require.NoError(t, err)
		back, err := val.Unpack(b, false)
		require.NoError(t, err, c.in)
		require.Equal(t, c.out, val.Hex(back), c.in)
	}
	// An empty payload is not an integer.
	_, err := val.Unpack([]byte{0xc7, 0x00, 0x01}, false)
	require.Error(t, err)
}

func TestDecimalPortablePayload(t *testing.T) {
	cases := []struct {
		coefficient string
		exponent    int32
		text        string
		hex         string
	}{
		{"0", 0, "0", "c705020000000000"},
		{"1", 0, "1", "c705020000000001"},
		{"10", -1, "1.0", "c70502000000010a"},
		{"110", -2, "1.10", "c70502000000026e"},
		{"12345", -2, "123.45", "c70602000000023039"},
		{"-1045", -3, "-1.045", "c7060200000003fbeb"},
		{"123567", -3, "123.567", "c707020000000301e2af"},
		{"-123567", -3, "-123.567", "c7070200000003fe1d51"},
		{"1", 3, "1000", "c70502fffffffd01"},
		{"-1", 3, "-1000", "c70502fffffffdff"},
		{"1", -10, "0.0000000001", "c705020000000a01"},
		{"123456789012345678901234567890", 0, "123456789012345678901234567890", "c7110200000000018ee90ff6c373e0ee4e3f0ad2"},
		{"-123456789012345678901234567890", -20, "-1234567890.1234567890123456789", "c7110200000014fe7116f0093c8c1f11b1c0f52e"},
	}
	for _, c := range cases {
		d := decimal.NewFromBigInt(bigFromString(t, c.coefficient), c.exponent)
		v := val.Decimal(d)
		require.Equal(t, c.hex, val.Hex(v), c.text)
		b, err := hex.DecodeString(c.hex)
		require.NoError(t, err)
		back, err := val.Unpack(b, false)
		require.NoError(t, err, c.text)
		require.True(t, v.Equal(back), "%s: %s != %s", c.text, v, back)
		bd := back.(val.Number).Decimal()
		require.Equal(t, c.exponent, bd.Exponent(), c.text)
		require.Equal(t, c.coefficient, bd.Coefficient().String(), c.text)
	}
	// Short payloads and empty unscaled integers are rejected.
	for _, in := range []string{"d6020000000a", "c705020000000a"[:12], "c7040200000002"} {
		b, err := hex.DecodeString(in)
		require.NoError(t, err)
		_, err = val.Unpack(b, false)
		require.Error(t, err, in)
	}
}

func TestDecimalScaleLimit(t *testing.T) {
	saved := val.MaxDecimalExponent
	defer func() { val.MaxDecimalExponent = saved }()
	val.MaxDecimalExponent = 100

	ok, err := hex.DecodeString("c70502000000640a") // scale 100
	require.NoError(t, err)
	_, err = val.Unpack(ok, false)
	require.NoError(t, err)
	tooBig, err := hex.DecodeString("c70502000000650a") // scale 101
	require.NoError(t, err)
	_, err = val.Unpack(tooBig, false)
	require.Error(t, err)
	tooSmall, err := hex.DecodeString("c70502ffffff9b0a") // scale -101
	require.NoError(t, err)
	_, err = val.Unpack(tooSmall, false)
	require.Error(t, err)

	val.MaxDecimalExponent = 0 // disabled
	_, err = val.Unpack(tooBig, false)
	require.NoError(t, err)
}

func TestLegacyBigNumbers(t *testing.T) {
	// Payloads written by value before v1.5.0: gob big.Int (version byte 2,
	// sign in bit 0) and shopspring MarshalBinary (int32 exponent + gob).
	legacyBigInt, err := hex.DecodeString("d501027b") // 123
	require.NoError(t, err)
	legacyDecimal, err := hex.DecodeString("d702fffffffd0201e2af") // 123.567
	require.NoError(t, err)

	// By default they are read in the portable layout, which gives other
	// numbers: there is no sniffing.
	v, err := val.Unpack(legacyBigInt, false)
	require.NoError(t, err)
	require.Equal(t, "635", v.(val.Number).BigInt().String()) // 0x027b

	saved := val.LegacyBigNumbers
	defer func() { val.LegacyBigNumbers = saved }()
	val.LegacyBigNumbers = true

	v, err = val.Unpack(legacyBigInt, false)
	require.NoError(t, err)
	require.Equal(t, "123", v.(val.Number).BigInt().String())
	v, err = val.Unpack(legacyDecimal, false)
	require.NoError(t, err)
	require.Equal(t, "123.567", v.(val.Number).Decimal().String())
	// Packing always writes the portable layout, so a migration is
	// Unpack then Pack.
	require.Equal(t, "c707020000000301e2af", val.Hex(v))

	x, err := val.UnpackLegacyBigInt([]byte{0x03, 0x7b})
	require.NoError(t, err)
	require.Equal(t, "-123", x.String())
	_, err = val.UnpackLegacyBigInt([]byte{0x09})
	require.Error(t, err)
	_, err = val.UnpackLegacyDecimal([]byte{0x00})
	require.Error(t, err)
}
