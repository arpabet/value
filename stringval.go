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

package val

import (
	"reflect"
	"encoding/base64"
	"strings"
	"strconv"
	"bytes"
	"github.com/pkg/errors"
)

const (
	jsonQuote = '"'
)

var Base64Prefix = "base64,"

type stringValue struct {
	dt 		StringType
	utf8 	string
	bytes 	[]byte
}

func (s stringValue) Equal(val Value) bool {
	if val == nil || val.Kind() != STRING {
		return false
	}
	o := val.(*stringValue)
	if s.dt != o.dt {
		return false
	}
	switch s.dt {
	case UTF8:
		return s.utf8 == o.utf8
	case RAW:
		return bytes.Compare(s.bytes, o.bytes) == 0
	}
	return false
}

func Utf8(val string) *stringValue {
	return &stringValue{
		dt: UTF8,
		utf8: val,
	}
}

func Raw(val []byte, copy bool) *stringValue {
	return &stringValue{
		dt: RAW,
		bytes: val,
	}
}

func ParseString(str string) *stringValue {
	if strings.HasPrefix(str, Base64Prefix) {
		raw, err := base64.RawStdEncoding.DecodeString(str[len(Base64Prefix):])
		if err == nil {
			return Raw(raw, false)
		}
	}
	return Utf8(str)
}

func (s stringValue) Kind() Kind {
	return STRING
}

func (s stringValue) Class() reflect.Type {
	return reflect.TypeOf((*stringValue)(nil)).Elem()
}

func (s stringValue) Len() int {
	switch s.dt {
	case UTF8:
		return len(s.utf8)
	case RAW:
		return len(s.bytes)
	default:
		return 0
	}
}

func (s stringValue) String() string {
	switch s.dt {
	case UTF8:
		return s.utf8
	case RAW:
		return Base64Prefix + base64.RawStdEncoding.EncodeToString(s.bytes)
	default:
		return ""
	}
}

func (s stringValue) Pack(p Packer) {
	switch s.dt {
	case UTF8:
		p.PackStr(s.utf8)
	case RAW:
		p.PackBin(s.bytes)
	default:
		p.PackNil()
	}
}

func (s stringValue) PrintJSON(out *strings.Builder) {
	switch s.dt {
	case UTF8:
		out.WriteString(strconv.Quote(s.utf8))
	case RAW:
		out.WriteRune(jsonQuote)
		out.WriteString(Base64Prefix)
		out.WriteString(base64.RawStdEncoding.EncodeToString(s.bytes))
		out.WriteRune(jsonQuote)
	default:
		out.WriteRune(jsonQuote)
		out.WriteRune(jsonQuote)
	}
}

func (s stringValue) MarshalJSON() ([]byte, error) {
	switch s.dt {
	case UTF8:
		return []byte(strconv.Quote(s.utf8)), nil
	case RAW:
		var out strings.Builder
		out.WriteRune(jsonQuote)
		out.WriteString(Base64Prefix)
		out.WriteString(base64.RawStdEncoding.EncodeToString(s.bytes))
		out.WriteRune(jsonQuote)
		return []byte(out.String()), nil
	default:
		return nil, errors.New("unknown string type")
	}
}

func (s stringValue) MarshalBinary() ([]byte, error) {
	buf := bytes.Buffer{}
	p := MessagePacker(&buf)
	s.Pack(p)
	return buf.Bytes(), p.Error()
}

func (s stringValue) Type() StringType {
	return s.dt
}

func (s stringValue) Utf8() string {
	switch s.dt {
	case UTF8:
		return s.utf8
	case RAW:
		return string(s.bytes)
	default:
		return ""
	}
}

func (s stringValue) Raw() []byte {
	switch s.dt {
	case UTF8:
		return []byte(s.utf8)
	case RAW:
		return s.bytes
	default:
		return []byte{}
	}
}
