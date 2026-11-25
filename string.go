/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */


package value

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)


const (
	jsonQuote = '"'
)

/*
quoteJSON writes s as a JSON string, in the rendering of Jsonify and PrintJSON:
a printable character as it is, any other escaped. It is strconv.Quote
wherever that gives JSON, and JSON where it does not. A Go string literal has
escapes JSON lacks, so \a, \v and \xNN of a control character are written
\u00NN, and a character beyond the basic plane that is not printable is
written as its surrogate pair in place of \UNNNNNNNN. A byte that is not
UTF-8 has no JSON form and is written as U+FFFD, as encoding/json writes it.
*/
func quoteJSON(out *strings.Builder, s string) {
	out.WriteByte(jsonQuote)
	for i := 0; i < len(s); {
		c := s[i]
		if c < utf8.RuneSelf {
			i++
			switch {
			case c == jsonQuote || c == '\\':
				out.WriteByte('\\')
				out.WriteByte(c)
			case c >= 0x20 && c != 0x7f:
				out.WriteByte(c)
			case c == '\b':
				out.WriteString(`\b`)
			case c == '\f':
				out.WriteString(`\f`)
			case c == '\n':
				out.WriteString(`\n`)
			case c == '\r':
				out.WriteString(`\r`)
			case c == '\t':
				out.WriteString(`\t`)
			default:
				escapeJSON(out, rune(c))
			}
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == utf8.RuneError && size == 1:
			escapeJSON(out, utf8.RuneError)
		case strconv.IsPrint(r):
			out.WriteString(s[i : i+size])
		case r < 0x10000:
			escapeJSON(out, r)
		default:
			high, low := utf16.EncodeRune(r)
			escapeJSON(out, high)
			escapeJSON(out, low)
		}
		i += size
	}
	out.WriteByte(jsonQuote)
}

// escapeJSON writes the \uNNNN escape of a UTF-16 code unit.
func escapeJSON(out *strings.Builder, unit rune) {
	out.WriteString(`\u`)
	for shift := 12; shift >= 0; shift -= 4 {
		out.WriteByte(hexDigits[unit>>uint(shift)&0xf])
	}
}

var Base64Prefix = "base64,"

type uft8String string
type rawString []byte

var uft8StringClass = reflect.TypeOf((*uft8String)(nil)).Elem()
var rawStringClass = reflect.TypeOf((*rawString)(nil)).Elem()

var EmptyString = uft8String("")

func Utf8(val string) String {
	return uft8String(val)
}

func Str(val string) String {
	return uft8String(val)
}

func Stringf(format string, args... interface{}) String {
	return uft8String(fmt.Sprintf(format, args...))
}

func (s uft8String) Type() StringType {
	return UTF8
}

func (s uft8String) Kind() Kind {
	return STRING
}

func (s uft8String) Class() reflect.Type {
	return uft8StringClass
}

func (s uft8String) Object() interface{} {
	return string(s)
}

func (s uft8String) String() string {
	return string(s)
}

func (s uft8String) Pack(p Packer) {
	p.PackStr(string(s))
}

func (s uft8String) PrintJSON(out *strings.Builder) {
	quoteJSON(out, string(s))
}

func (s uft8String) MarshalJSON() ([]byte, error) {
	var out strings.Builder
	quoteJSON(&out, string(s))
	return []byte(out.String()), nil
}

func (s uft8String) MarshalBinary() ([]byte, error) {
	buf := bytes.Buffer{}
	p := MessagePacker(&buf)
	s.Pack(p)
	return buf.Bytes(), p.Error()
}

func (s uft8String) Equal(val Value) bool {
	if val == nil || val.Kind() != STRING {
		return false
	}
	return string(s) == val.String()
}

func (s uft8String) Len() int {
	return len(s)
}

func (s uft8String) Utf8() string {
	return string(s)
}

func (s uft8String) Raw() []byte {
	return []byte(s)
}

func Raw(val []byte, copyFlag bool) String {
	if copyFlag {
		dst := make([]byte, len(val))
		copy(dst, val)
		return rawString(dst)
	} else {
		return rawString(val)
	}
}

func (s rawString) Type() StringType {
	return RAW
}

func (s rawString) Kind() Kind {
	return STRING
}

func (s rawString) Class() reflect.Type {
	return rawStringClass
}

func (s rawString) Object() interface{} {
	return []byte(s)
}

func (s rawString) String() string {
	return Base64Prefix + base64.RawStdEncoding.EncodeToString(s)
}

func (s rawString) Pack(p Packer) {
	p.PackBin(s)
}

func (s rawString) PrintJSON(out *strings.Builder) {
	out.WriteRune(jsonQuote)
	out.WriteString(Base64Prefix)
	out.WriteString(base64.RawStdEncoding.EncodeToString(s))
	out.WriteRune(jsonQuote)
}

func (s rawString) MarshalJSON() ([]byte, error) {
	var out strings.Builder
	out.WriteRune(jsonQuote)
	out.WriteString(Base64Prefix)
	out.WriteString(base64.RawStdEncoding.EncodeToString(s))
	out.WriteRune(jsonQuote)
	return []byte(out.String()), nil
}

func (s rawString) MarshalBinary() ([]byte, error) {
	buf := bytes.Buffer{}
	p := MessagePacker(&buf)
	s.Pack(p)
	return buf.Bytes(), p.Error()
}

func (s rawString) Equal(val Value) bool {
	if val == nil || val.Kind() != STRING {
		return false
	}
	o := val.(String)
	return bytes.Equal(s, o.Raw())
}

func (s rawString) Len() int {
	return len(s)
}

func (s rawString) Utf8() string {
	return string(s)
}

func (s rawString) Raw() []byte {
	return s
}

func ParseString(str string) String {
	if strings.HasPrefix(str, Base64Prefix) {
		raw, err := base64.RawStdEncoding.DecodeString(str[len(Base64Prefix):])
		if err == nil {
			return Raw(raw, false)
		}
	}
	return Utf8(str)
}
