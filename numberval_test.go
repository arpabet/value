/*
 *
 * Copyright 2019-present Alexander Shvid
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package genval_test

import (
	"testing"
	"github.com/shvid/genval"
	"github.com/stretchr/testify/require"
	"math"
	"encoding/json"
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
	require.Equal(t, "0", genval.Json(b))
	require.Equal(t, "0", b.String())

	b = genval.Long(1)

	require.Equal(t, genval.NUMBER, b.Kind())
	require.Equal(t, genval.LONG, b.Type())
	require.Equal(t, "genval.numberValue", b.Class().String())
	require.Equal(t, "01", genval.Hex(b))
	require.Equal(t, "1", genval.Json(b))
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
	require.Equal(t, "0", genval.Json(b))
	require.Equal(t, "0", b.String())
	require.Equal(t, int64(0), b.Long())

	b = genval.Double(1)
	require.Equal(t, genval.NUMBER, b.Kind())
	require.Equal(t, genval.DOUBLE, b.Type())
	require.Equal(t, "genval.numberValue", b.Class().String())
	require.Equal(t, "cb3ff0000000000000", genval.Hex(b))
	require.Equal(t, "1", genval.Json(b))
	require.Equal(t, "1", b.String())
	require.Equal(t, int64(1), b.Long())

	b = genval.Double(123456789)
	require.Equal(t, genval.NUMBER, b.Kind())
	require.Equal(t, genval.DOUBLE, b.Type())
	require.Equal(t, "genval.numberValue", b.Class().String())
	require.Equal(t, "cb419d6f3454000000", genval.Hex(b))
	require.Equal(t, "1.23456789e+08", genval.Json(b))
	require.Equal(t, "1.23456789e+08", b.String())
	require.Equal(t, int64(123456789), b.Long())

	b = genval.Double(-123456789)
	require.Equal(t, genval.NUMBER, b.Kind())
	require.Equal(t, genval.DOUBLE, b.Type())
	require.Equal(t, "genval.numberValue", b.Class().String())
	require.Equal(t, "cbc19d6f3454000000", genval.Hex(b))
	require.Equal(t, "-1.23456789e+08", genval.Json(b))
	require.Equal(t, "-1.23456789e+08", b.String())
	require.Equal(t, int64(-123456789), b.Long())

}

func TestParseNumber(t *testing.T) {

	b := genval.ParseNumber("0")

	require.Equal(t, genval.NUMBER, b.Kind())
	require.Equal(t, genval.LONG, b.Type())
	require.Equal(t, "genval.numberValue", b.Class().String())
	require.Equal(t, "00", genval.Hex(b))
	require.Equal(t, "0", genval.Json(b))
	require.Equal(t, "0", b.String())

	b = genval.ParseNumber("123")

	require.Equal(t, genval.NUMBER, b.Kind())
	require.Equal(t, genval.LONG, b.Type())
	require.Equal(t, "genval.numberValue", b.Class().String())
	require.Equal(t, "7b", genval.Hex(b))
	require.Equal(t, "123", genval.Json(b))
	require.Equal(t, "123", b.String())

	b = genval.ParseNumber("-123")

	require.Equal(t, genval.NUMBER, b.Kind())
	require.Equal(t, genval.LONG, b.Type())
	require.Equal(t, "genval.numberValue", b.Class().String())
	require.Equal(t, "d085", genval.Hex(b))
	require.Equal(t, "-123", genval.Json(b))
	require.Equal(t, "-123", b.String())

	b = genval.ParseNumber("123.45")

	require.Equal(t, genval.NUMBER, b.Kind())
	require.Equal(t, genval.DOUBLE, b.Type())
	require.Equal(t, "genval.numberValue", b.Class().String())
	require.Equal(t, "cb405edccccccccccd", genval.Hex(b))
	require.Equal(t, "123.45", genval.Json(b))
	require.Equal(t, "123.45", b.String())

	b = genval.ParseNumber("-123.45")

	require.Equal(t, genval.NUMBER, b.Kind())
	require.Equal(t, genval.DOUBLE, b.Type())
	require.Equal(t, "genval.numberValue", b.Class().String())
	require.Equal(t, "cbc05edccccccccccd", genval.Hex(b))
	require.Equal(t, "-123.45", genval.Json(b))
	require.Equal(t, "-123.45", b.String())

	b = genval.ParseNumber("123456789.123456789")

	require.Equal(t, genval.NUMBER, b.Kind())
	require.Equal(t, genval.DOUBLE, b.Type())
	require.Equal(t, "genval.numberValue", b.Class().String())
	require.Equal(t, "cb419d6f34547e6b75", genval.Hex(b))
	require.Equal(t, "1.2345678912345679e+08", genval.Json(b))
	require.Equal(t, "1.2345678912345679e+08", b.String())

	c := genval.ParseNumber("1.2345678912345679e+08")
	DoubleEqual(t, b.Double(), c.Double())

	b = genval.ParseNumber("-123456789.123456789")

	require.Equal(t, genval.NUMBER, b.Kind())
	require.Equal(t, genval.DOUBLE, b.Type())
	require.Equal(t, "genval.numberValue", b.Class().String())
	require.Equal(t, "cbc19d6f34547e6b75", genval.Hex(b))
	require.Equal(t, "-1.2345678912345679e+08", genval.Json(b))
	require.Equal(t, "-1.2345678912345679e+08", b.String())

	c = genval.ParseNumber("-1.2345678912345679e+08")
	DoubleEqual(t, b.Double(), c.Double())

}

func TestParseNaN(t *testing.T) {

	b := genval.ParseNumber("not a number")

	require.Equal(t, genval.NUMBER, b.Kind())
	require.Equal(t, genval.DOUBLE, b.Type())
	require.True(t, math.IsNaN(b.Double()))
	require.Equal(t, int64(0), b.Long())
	require.Equal(t, "genval.numberValue", b.Class().String())
	require.Equal(t, "cb7ff8000000000001", genval.Hex(b))
	require.Equal(t, "null", genval.Json(b))
	require.Equal(t, "NaN", b.String())

}

func TestAddNumber(t *testing.T) {

	a := genval.ParseNumber("3")
	b := genval.ParseNumber("2")

	c := a.Add(b)

	require.Equal(t, genval.NUMBER, c.Kind())
	require.Equal(t, genval.LONG, c.Type())

	require.Equal(t, int64(3), a.Long())
	require.Equal(t, int64(2), b.Long())
	require.Equal(t, int64(5), c.Long())

}

func TestSubtractNumber(t *testing.T) {

	a := genval.ParseNumber("3")
	b := genval.ParseNumber("2")

	c := a.Subtract(b)

	require.Equal(t, genval.NUMBER, c.Kind())
	require.Equal(t, genval.LONG, c.Type())

	require.Equal(t, int64(3), a.Long())
	require.Equal(t, int64(2), b.Long())
	require.Equal(t, int64(1), c.Long())

}

func TestAddFloatNumber(t *testing.T) {

	a := genval.ParseNumber("3.3")
	b := genval.ParseNumber("2.2")

	c := a.Add(b)

	require.Equal(t, genval.NUMBER, c.Kind())
	require.Equal(t, genval.DOUBLE, c.Type())

	DoubleEqual(t, float64(3.3), a.Double())
	DoubleEqual(t, float64(2.2), b.Double())
	DoubleEqual(t, float64(5.5), c.Double())

}

func TestSubtractFloatNumber(t *testing.T) {

	a := genval.ParseNumber("3.3")
	b := genval.ParseNumber("2.2")

	c := a.Subtract(b)

	require.Equal(t, genval.NUMBER, c.Kind())
	require.Equal(t, genval.DOUBLE, c.Type())

	DoubleEqual(t, float64(3.3), a.Double())
	DoubleEqual(t, float64(2.2), b.Double())
	DoubleEqual(t, float64(1.1), c.Double())

}

func TestAddNaN(t *testing.T) {

	a := genval.ParseNumber("3.3")
	b := genval.ParseNumber("NaN")

	c := a.Add(b)

	require.Equal(t, genval.NUMBER, c.Kind())
	require.Equal(t, genval.DOUBLE, c.Type())

	DoubleEqual(t, float64(3.3), a.Double())
	require.True(t, math.IsNaN(b.Double()))
	require.True(t, math.IsNaN(c.Double()))

}

func TestSubtractNaN(t *testing.T) {

	a := genval.ParseNumber("3.3")
	b := genval.ParseNumber("NaN")

	c := a.Subtract(b)

	require.Equal(t, genval.NUMBER, c.Kind())
	require.Equal(t, genval.DOUBLE, c.Type())

	DoubleEqual(t, float64(3.3), a.Double())
	require.True(t, math.IsNaN(b.Double()))
	require.True(t, math.IsNaN(c.Double()))

}

func TestAddNaNBoth(t *testing.T) {

	a := genval.ParseNumber("NaN")
	b := genval.ParseNumber("NaN")

	c := a.Add(b)

	require.Equal(t, genval.NUMBER, c.Kind())
	require.Equal(t, genval.DOUBLE, c.Type())

	require.True(t, math.IsNaN(a.Double()))
	require.True(t, math.IsNaN(b.Double()))
	require.True(t, math.IsNaN(c.Double()))

}

func TestSubtractNaNBoth(t *testing.T) {

	a := genval.ParseNumber("NaN")
	b := genval.ParseNumber("NaN")

	c := a.Subtract(b)

	require.Equal(t, genval.NUMBER, c.Kind())
	require.Equal(t, genval.DOUBLE, c.Type())

	require.True(t, math.IsNaN(a.Double()))
	require.True(t, math.IsNaN(b.Double()))
	require.True(t, math.IsNaN(c.Double()))

}

func DoubleEqual(t *testing.T, left, right float64) {
	require.True(t, math.Abs(left - right) < 0.00001)
}

type testNumberStruct struct {
	N genval.Number
}

func TestNumberMarshal(t *testing.T) {

	b := genval.Long(123)

	j, _ := b.MarshalJSON()
	require.Equal(t, "123", string(j))

	bin, _ := b.MarshalBinary()
	require.Equal(t, []byte{0x7b}, bin)

	b = genval.Double(1.23)

	j, _ = b.MarshalJSON()
	require.Equal(t, "1.23", string(j))

	bin, _ = b.MarshalBinary()
	require.Equal(t, []byte{0xcb, 0x3f, 0xf3, 0xae, 0x14, 0x7a, 0xe1, 0x47, 0xae}, bin)

	s := &testNumberStruct{genval.Long(123)}

	j, _ = json.Marshal(s)
	require.Equal(t, "{\"N\":123}", string(j))

}

func TestPackLong(t *testing.T) {

	for num, _ := range testLongMap {

		b := genval.Long(num)

		mp, err := genval.Pack(b)
		if err != nil {
			t.Errorf("pack fail %v", err)
		}
		c, err := genval.Unpack(mp, false)
		if err != nil {
			t.Errorf("unpack fail %v", err)
		}

		require.Truef(t, b.Equal(c), "expected: %v, actual: %v", b, c)
	}

}