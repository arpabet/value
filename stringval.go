package genval

import (
	"reflect"
	"encoding/base64"
	"strings"
	"strconv"
)

const (
	jsonQuote = "\""
	base64Prefix = "base64!"
)

var jsonQuoteByte = byte(jsonQuote[0])

type stringValue struct {
	dt 		StringType
	utf8 	string
	bytes 	[]byte
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
	if strings.HasPrefix(str, base64Prefix) {
		raw, err := base64.RawStdEncoding.DecodeString(str[len(base64Prefix):])
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
		return base64Prefix + base64.RawStdEncoding.EncodeToString(s.bytes)
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
		return jsonQuote + base64Prefix + base64.RawStdEncoding.EncodeToString(s.bytes) + jsonQuote
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
