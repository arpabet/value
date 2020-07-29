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
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	jsonQuote = '"'
)

var Base64Prefix = "base64,"

type uft8String string
type rawString []byte

func Utf8(val string) String {
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
	return reflect.TypeOf((*uft8String)(nil)).Elem()
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
	out.WriteString(strconv.Quote(string(s)))
}

func (s uft8String) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(string(s))), nil
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
	return reflect.TypeOf((*rawString)(nil)).Elem()
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
	return bytes.Compare(s, o.Raw()) == 0
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
