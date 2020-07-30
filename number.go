/*
 *
 * Copyright 2020-present Arpabet, Inc.
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

package value

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"github.com/shopspring/decimal"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"
)


/**
	@author Alex Shvid
*/

var DecimalExpDelim = byte('x')
var DecimalExpDelimStr = "x"

var PrecisionLevel = 0.00001

type longNumber int64
type doubleNumber float64

type bigIntNumber struct {
	*big.Int
}

type decimalNumber decimal.Decimal


func Long(val int64) Number {
	return longNumber(val)
}

func Double(val float64) Number {
	return doubleNumber(val)
}

func BigInt(val *big.Int) Number {
	if val == nil {
		val = new(big.Int)
	}
	return bigIntNumber{val}
}

func Decimal(dec decimal.Decimal) Number {
	return decimalNumber(dec)
}

func Nan() Number {
	return doubleNumber(math.NaN())
}

func ParseNumber(str string) Number {

	if len(str) == 0 {
		return Long(0)
	}

	if str == "null" {
		return Nan()
	}

	if strings.IndexByte(str, '.') != -1 {
		double, err := strconv.ParseFloat(str, 64)
		if err == nil {
			return Double(double)
		}
		return Nan()
	}

	if hasHexPrefix(str) {
		if val, err := parseHexNumber(str); err == nil {
			return val
		}
	}

	long, err := strconv.ParseInt(str, 10, 64)
	if err == nil {
		return Long(long)
	} else {
		double, err := strconv.ParseFloat(str, 64)
		if err == nil {
			return Double(double)
		}
	}

	return Nan()

}

func parseHexNumber(s string) (Number, error) {
	neg := false
	if len(s) >= 1 && s[0] == '-' {
		s = s[1:]
		neg = true
	}
	if len(s) >= 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		s = s[2:]  // 0x
	}
	exp := int32(0)
	expIdx := strings.IndexByte(s, DecimalExpDelim)
	if expIdx != -1 {
		var err error
		exp, err = parseExp(s[expIdx+1:])
		if err != nil {
			return nil, err
		}
		s = s[:expIdx]
	}

	if len(s) % 2 == 1 {
		s = "0" + s
	}

	val, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	mantissa := new(big.Int).SetBytes(val)
	if neg {
		mantissa = new(big.Int).Neg(mantissa)
	}

	if exp == 0 {
		return BigInt(mantissa), nil
	} else {
		dec := decimal.NewFromBigInt(mantissa, exp)
		return Decimal(dec), nil
	}

}

func hasHexPrefix(s string) bool {
	if len(s) >= 1 && s[0] == '-' {
		s = s[1:]
	}
	return len(s) >= 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X')
}

func (n longNumber) Type() NumberType {
	return LONG
}

func (n doubleNumber) Type() NumberType {
	return DOUBLE
}

func (n bigIntNumber) Type() NumberType {
	return BIGINT
}

func (n decimalNumber) Type() NumberType {
	return DECIMAL
}

func (n longNumber) Kind() Kind {
	return NUMBER
}

func (n doubleNumber) Kind() Kind {
	return NUMBER
}

func (n bigIntNumber) Kind() Kind {
	return NUMBER
}

func (n decimalNumber) Kind() Kind {
	return NUMBER
}

func (n longNumber) Class() reflect.Type {
	return reflect.TypeOf((*longNumber)(nil)).Elem()
}

func (n doubleNumber) Class() reflect.Type {
	return reflect.TypeOf((*doubleNumber)(nil)).Elem()
}

func (n bigIntNumber) Class() reflect.Type {
	return reflect.TypeOf((*bigIntNumber)(nil)).Elem()
}

func (n decimalNumber) Class() reflect.Type {
	return reflect.TypeOf((*decimalNumber)(nil)).Elem()
}

func (n longNumber) String() string {
	return strconv.FormatInt(int64(n), 10)
}

func (n doubleNumber) String() string {
	d := float64(n)
	if math.IsNaN(d) {
		return "NaN"
	} else {
		return strconv.FormatFloat(d, 'f', -1, 64)
	}
}

func (n bigIntNumber) String() string {
	return formatBigInt(n.Int)
}

func (n decimalNumber) String() string {
	return formatDecimal(decimal.Decimal(n))
}

func formatBigInt(val *big.Int) string {
	s := hex.EncodeToString(val.Bytes())
	if s == "" {
		s = "0"
	}
	if val.Sign() >= 0 {
		return "0x" + s
	} else {
		return "-0x" + s
	}
}

func formatDecimal(dec decimal.Decimal) string {
	exp := dec.Exponent()
	val := dec.Coefficient()
	s := hex.EncodeToString(val.Bytes())
	if s == "" {
		s = "0"
	}
	if exp != 0 {
		s = s + DecimalExpDelimStr + formatExp(exp)
	}
	if val.Sign() >= 0 {
		return "0x" + s
	} else {
		return "-0x" + s
	}
}

func formatExp(exp int32) string {
	if exp == 0 {
		return ""
	}
	s := ""
	if exp < 0 {
		s = "-"
		exp = -exp
	}
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(exp))
	var slice []byte
	for i := 0; i < len(buf); i++ {
		if buf[i] != 0 {
			slice = buf[i:]
			break
		}
	}
	return s + hex.EncodeToString(slice)
}

func parseExp(s string) (int32, error) {
	neg := false
	if len(s) >= 1 && s[0] == '-' {
		s = s[1:]
		neg = true
	}

	if len(s) % 2 == 1 {
		s = "0" + s
	}

	bytes, err := hex.DecodeString(s)
	if err != nil {
		return 0, err
	}
	var buf [4]byte
	copy(buf[len(buf)-len(bytes):], bytes)
	val := int32(binary.BigEndian.Uint32(buf[:]))
	if neg {
		val = -val
	}
	return val, nil
}

func (n longNumber) Object() interface{} {
	return int64(n)
}

func (n doubleNumber) Object() interface{} {
	return float64(n)
}

func (n bigIntNumber) Object() interface{} {
	return n.Int
}

func (n decimalNumber) Object() interface{} {
	return decimal.Decimal(n)
}

func (n longNumber) Pack(p Packer) {
	p.PackLong(int64(n))
}

func (n doubleNumber) Pack(p Packer) {
	p.PackDouble(float64(n))
}

func (n bigIntNumber) Pack(p Packer) {
	b, _ := n.Int.GobEncode()
	p.PackExt(BigIntExt, b)
}

func (n decimalNumber) Pack(p Packer) {
	dec := decimal.Decimal(n)
	b, _ := dec.MarshalBinary()
	p.PackExt(DecimalExt, b)
}

func UnpackBigInt(data []byte) (*big.Int, error) {
	x := new(big.Int)
	err := x.GobDecode(data)
	return x, err
}

func UnpackDecimal(data []byte) (decimal.Decimal, error) {
	d := decimal.Decimal{}
	err := d.UnmarshalBinary(data)
	return d, err
}

func (n longNumber) PrintJSON(out *strings.Builder) {
	out.WriteString(n.String())
}

func (n doubleNumber) PrintJSON(out *strings.Builder) {
	d := float64(n)
	if math.IsNaN(d) {
		out.WriteString("null")
	} else {
		out.WriteString(strconv.FormatFloat(d, 'f', -1, 64))
	}
}

func (n bigIntNumber) PrintJSON(out *strings.Builder) {
	out.WriteRune(jsonQuote)
	out.WriteString(formatBigInt(n.Int))
	out.WriteRune(jsonQuote)
}

func (n decimalNumber) PrintJSON(out *strings.Builder) {
	out.WriteRune(jsonQuote)
	out.WriteString(formatDecimal(decimal.Decimal(n)))
	out.WriteRune(jsonQuote)
}

func (n longNumber) MarshalJSON() ([]byte, error) {
	return []byte(n.String()), nil
}

func (n doubleNumber) MarshalJSON() ([]byte, error) {
	d := float64(n)
	if math.IsNaN(d) {
		return []byte("null"), nil
	} else {
		return []byte(strconv.FormatFloat(d, 'f', -1, 64)), nil
	}
}

func (n bigIntNumber) MarshalJSON() ([]byte, error) {
	s := formatBigInt(n.Int)
	return []byte(strconv.Quote(s)), nil
}

func (n decimalNumber) MarshalJSON() ([]byte, error) {
	s := formatDecimal(decimal.Decimal(n))
	return []byte(strconv.Quote(s)), nil
}

func (n longNumber) MarshalBinary() ([]byte, error) {
	m := new(messageWriter) // must be in heap
	return m.WriteLong(int64(n)), nil
}

func (n doubleNumber) MarshalBinary() ([]byte, error) {
	m := new(messageWriter) // must be in heap
	return m.WriteDouble(float64(n)), nil
}

func (n bigIntNumber) MarshalBinary() ([]byte, error) {
	buf := bytes.Buffer{}
	p := MessagePacker(&buf)
	n.Pack(p)
	return buf.Bytes(), p.Error()
}

func (n decimalNumber) MarshalBinary() ([]byte, error) {
	buf := bytes.Buffer{}
	p := MessagePacker(&buf)
	n.Pack(p)
	return buf.Bytes(), p.Error()
}

func (n longNumber) IsNaN() bool {
	return false
}

func (n doubleNumber) IsNaN() bool {
	d := float64(n)
	return math.IsNaN(d)
}

func (n bigIntNumber) IsNaN() bool {
	return false
}

func (n decimalNumber) IsNaN() bool {
	return false
}

func (n longNumber) Long() int64 {
	return int64(n)
}

func (n doubleNumber) Long() int64 {
	d := float64(n)
	if math.IsNaN(d) {
		return 0
	} else {
		return int64(d)
	}
}

func (n bigIntNumber) Long() int64 {
	return n.Int.Int64()
}

func (n decimalNumber) Long() int64 {
	return decimal.Decimal(n).IntPart()
}

func (n longNumber) Double() float64 {
	return float64(n)
}

func (n doubleNumber) Double() float64 {
	return float64(n)
}

func (n bigIntNumber) Double() float64 {
	f := new(big.Float)
	f.SetInt(n.Int)
	d, _ := f.Float64()
	return d
}

func (n decimalNumber) Double() float64 {
	v, _ := decimal.Decimal(n).Float64()
	return v
}

func (n longNumber) BigInt() *big.Int {
	return big.NewInt(int64(n))
}

func (n doubleNumber) BigInt() *big.Int {
	d := float64(n)
	if math.IsNaN(d) {
		return big.NewInt(0)
	}
	f := new(big.Float)
	f.SetFloat64(d)
	z, _ := f.Int(nil)
	return z
}

func (n bigIntNumber) BigInt() *big.Int {
	return n.Int
}

func (n decimalNumber) BigInt() *big.Int {
	return decimal.Decimal(n).Floor().Coefficient()
}

func (n longNumber) Decimal() decimal.Decimal {
	return decimal.NewFromInt(int64(n))
}

func (n doubleNumber) Decimal() decimal.Decimal {
	d := float64(n)
	if math.IsNaN(d) {
		return decimal.NewFromInt(0)
	}
	return decimal.NewFromFloat(d)
}

func (n bigIntNumber) Decimal() decimal.Decimal {
	return decimal.NewFromBigInt(n.Int, 0)
}

func (n decimalNumber) Decimal() decimal.Decimal {
	return decimal.Decimal(n)
}

func (n longNumber) Add(other Number) Number {
	return Long(n.Long() + other.Long())
}

func (n doubleNumber) Add(other Number) Number {
	left := float64(n)
	right := other.Double()
	if math.IsNaN(left) || math.IsNaN(right) {
		return Nan()
	}
	return Double(left + right)
}

func (n bigIntNumber) Add(other Number) Number {
	z := new(big.Int)
	z.Add(n.Int, other.BigInt())
	return BigInt(z)
}

func (n decimalNumber) Add(other Number) Number {
	dec := decimal.Decimal(n)
	return Decimal(dec.Add(other.Decimal()))
}

func (n longNumber) Subtract(other Number) Number {
	return Long(n.Long() - other.Long())
}

func (n doubleNumber) Subtract(other Number) Number {
	left := float64(n)
	right := other.Double()
	if math.IsNaN(left) || math.IsNaN(right) {
		return Nan()
	}
	return Double(left - right)
}

func (n bigIntNumber) Subtract(other Number) Number {
	z := new(big.Int)
	z.Sub(n.Int, other.BigInt())
	return BigInt(z)
}

func (n decimalNumber) Subtract(other Number) Number {
	dec := decimal.Decimal(n)
	return Decimal(dec.Sub(other.Decimal()))
}

func (n longNumber) Equal(val Value) bool {
	if val == nil || val.Kind() != NUMBER {
		return false
	}
	other := val.(Number)
	return n.Long() == other.Long()
}

func (n doubleNumber) Equal(val Value) bool {
	if val == nil || val.Kind() != NUMBER {
		return false
	}
	other := val.(Number)
	if n.IsNaN() || other.IsNaN() {
		return false
	}
	return  math.Abs(n.Double() - other.Double()) < PrecisionLevel
}

func (n bigIntNumber) Equal(val Value) bool {
	if val == nil || val.Kind() != NUMBER {
		return false
	}
	other := val.(Number)
	return n.Int.Cmp(other.BigInt()) == 0
}

func (n decimalNumber) Equal(val Value) bool {
	if val == nil || val.Kind() != NUMBER {
		return false
	}
	other := val.(Number)
	return n.Decimal().Cmp(other.Decimal()) == 0
}

