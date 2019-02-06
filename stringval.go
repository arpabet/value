package genval

import (
	"reflect"
)

const (
	jsonQuote = "\""
)

var jsonQuoteByte = byte(jsonQuote[0])

type StringValue struct {
	dt 		StringType
	utf8 	string
	bytes 	[]byte
}

func Utf8(val string) *StringValue {
	return &StringValue{
		dt: UTF8,
		utf8: val,
	}
}

func Raw(val []byte, copy bool) *StringValue {
	return &StringValue{
		dt: RAW,
		bytes: val,
	}
}

func (s StringValue) Kind() Kind {
	return STRING
}

func (s StringValue) Class() reflect.Type {
	return reflect.TypeOf((*StringValue)(nil)).Elem()
}

func (s StringValue) Len() int {
	switch s.dt {
	case UTF8:
		return len(s.utf8)
	case RAW:
		return len(s.bytes)
	default:
		return 0
	}
}

func (s StringValue) String() string {
	switch s.dt {
	case UTF8:
		return s.utf8
	case RAW:
		return string(s.bytes)
	default:
		return ""
	}
}

func (s StringValue) Pack(p Packer) {
	switch s.dt {
	case UTF8:
		p.PackString(s.utf8)
	case RAW:
		p.PackBytes(s.bytes)
	default:
		p.PackNil()
	}
}

func (s StringValue) Json() string {
	switch s.dt {
	case UTF8:
		return jsonQuote + s.utf8 + jsonQuote
	case RAW:
		l := len(s.bytes)
		b := make([]byte, 1 + l + 1)
		b[0] = jsonQuoteByte
		copy(b[1:], s.bytes)
		b[l] = jsonQuoteByte
		return string(b)
	default:
		return jsonQuote + jsonQuote
	}
}

func (s StringValue) Type() StringType {
	return s.dt
}

func (s StringValue) Utf8() string {
	switch s.dt {
	case UTF8:
		return s.utf8
	case RAW:
		return string(s.bytes)
	default:
		return ""
	}
}

func (s StringValue) Bytes() []byte {
	switch s.dt {
	case UTF8:
		return []byte(s.utf8)
	case RAW:
		return s.bytes
	default:
		return []byte{}
	}
}
