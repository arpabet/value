package genval

import (
	"reflect"
	"encoding/base64"
	"strings"
	"strconv"
	"bytes"
)

const (
	jsonQuote = "\""
)

var jsonQuoteByte = byte(jsonQuote[0])

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
		p.PackString(s.utf8)
	case RAW:
		p.PackBytes(s.bytes)
	default:
		p.PackNil()
	}
}

func (s stringValue) Json() string {
	switch s.dt {
	case UTF8:
		return strconv.Quote(s.utf8)
	case RAW:
		return jsonQuote + Base64Prefix + base64.RawStdEncoding.EncodeToString(s.bytes) + jsonQuote
	default:
		return jsonQuote + jsonQuote
	}
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
