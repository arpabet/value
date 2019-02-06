package genval_test

import (
	"testing"
	"github.com/shvid/genval"
	"github.com/stretchr/testify/require"
	"math"
)

var testLongMap = map[int64]string {

	-9223372036854775808: "d38000000000000000",
	-9223372036854775807: "d38000000000000001",
	-9223372036854775806: "d38000000000000002",
	-2147483651: "d3ffffffff7ffffffd",
	-2147483650: "d3ffffffff7ffffffe",
	-2147483649: "d3ffffffff7fffffff",
	-2147483648: "d280000000",
	-2147483647: "d280000001",
	-2147483646: "d280000002",
	-32771: "d2ffff7ffd",
	-32770: "d2ffff7ffe",
	-32769: "d2ffff7fff",
	-32768: "d18000",
	-32767: "d18001",
	-131: "d1ff7d",
	-130: "d1ff7e",
	-129: "d1ff7f",
	-128: "d080",
	-127: "d081",
	-34: "d0de",
	-33: "d0df",
	-32: "e0",
	-31: "e1",
	0: "00",
	1: "01",
	126: "7e",
	127: "7f",
	128: "cc80",
	129: "cc81",
	130: "cc82",
	32765: "cd7ffd",
	32766: "cd7ffe",
	32767: "cd7fff",
	32768: "cd8000",
	32769: "cd8001",
	32770: "cd8002",
	2147483645: "ce7ffffffd",
	2147483646: "ce7ffffffe",
	2147483647: "ce7fffffff",
	2147483648: "ce80000000",
	2147483649: "ce80000001",
	2147483650: "ce80000002",
	4294967296: "cf0000000100000000",
	4294967297: "cf0000000100000001",
	4294967298: "cf0000000100000002",

}

func TestLongNumber(t *testing.T) {

	b := genval.Long(0)

	require.Equal(t, genval.NUMBER, b.Kind())
	require.Equal(t, genval.LONG, b.Type())
	require.Equal(t, "genval.numberValue", b.Class().String())
	require.Equal(t, "00", genval.Hex(b))
	require.Equal(t, "0", b.Json())
	require.Equal(t, "0", b.String())

	b = genval.Long(1)

	require.Equal(t, genval.NUMBER, b.Kind())
	require.Equal(t, genval.LONG, b.Type())
	require.Equal(t, "genval.numberValue", b.Class().String())
	require.Equal(t, "01", genval.Hex(b))
	require.Equal(t, "1", b.Json())
	require.Equal(t, "1", b.String())

	for num, hex := range testLongMap {
		b = genval.Long(num)
		require.True(t, math.Abs(float64(num) - b.Double()) < 0.0001)
		require.Equal(t, hex, genval.Hex(b))
	}

}

func TestDoubleNumber(t *testing.T) {

	b := genval.Double(0)
	require.Equal(t, genval.NUMBER, b.Kind())
	require.Equal(t, genval.DOUBLE, b.Type())
	require.Equal(t, "genval.numberValue", b.Class().String())
	require.Equal(t, "cb0000000000000000", genval.Hex(b))
	require.Equal(t, "0", b.Json())
	require.Equal(t, "0", b.String())
	require.Equal(t, int64(0), b.Long())

	b = genval.Double(1)
	require.Equal(t, genval.NUMBER, b.Kind())
	require.Equal(t, genval.DOUBLE, b.Type())
	require.Equal(t, "genval.numberValue", b.Class().String())
	require.Equal(t, "cb3ff0000000000000", genval.Hex(b))
	require.Equal(t, "1", b.Json())
	require.Equal(t, "1", b.String())
	require.Equal(t, int64(1), b.Long())

	b = genval.Double(123456789)
	require.Equal(t, genval.NUMBER, b.Kind())
	require.Equal(t, genval.DOUBLE, b.Type())
	require.Equal(t, "genval.numberValue", b.Class().String())
	require.Equal(t, "cb419d6f3454000000", genval.Hex(b))
	require.Equal(t, "1.23456789e+08", b.Json())
	require.Equal(t, "1.23456789e+08", b.String())
	require.Equal(t, int64(123456789), b.Long())

	b = genval.Double(-123456789)
	require.Equal(t, genval.NUMBER, b.Kind())
	require.Equal(t, genval.DOUBLE, b.Type())
	require.Equal(t, "genval.numberValue", b.Class().String())
	require.Equal(t, "cbc19d6f3454000000", genval.Hex(b))
	require.Equal(t, "-1.23456789e+08", b.Json())
	require.Equal(t, "-1.23456789e+08", b.String())
	require.Equal(t, int64(-123456789), b.Long())

}