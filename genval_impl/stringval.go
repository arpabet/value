package genval_impl

import (
	"github.com/shvid/genval"
	"reflect"
)

const (
	jsonQuote = "\""
)

var jsonQuoteByte = byte(jsonQuote[0])

type StringVal struct {
	dt 		genval.StringType
	utf8 	string
	bytes 	[]byte
}

func (s StringVal) Kind() genval.Kind {
	return genval.StringVal
}

func (s StringVal) Class() reflect.Type {
	return reflect.TypeOf((*StringVal)(nil)).Elem()
}

func (s StringVal) Len() int {
	switch s.dt {
	case genval.UTF8:
		return len(s.utf8)
	case genval.Bytes:
		return len(s.bytes)
	default:
		return 0
	}
}

func (s StringVal) String() string {
	switch s.dt {
	case genval.UTF8:
		return s.utf8
	case genval.Bytes:
		return string(s.bytes)
	default:
		return ""
	}
}

func (s StringVal) Pack(p genval.Packer) {
	switch s.dt {
	case genval.UTF8:
		p.PackString(s.utf8)
	case genval.Bytes:
		p.PackBytes(s.bytes)
	default:
		p.PackNil()
	}
}

func (s StringVal) Json() string {
	switch s.dt {
	case genval.UTF8:
		return jsonQuote + s.utf8 + jsonQuote
	case genval.Bytes:
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

func (s StringVal) Type() genval.StringType {
	return s.dt
}

func (s StringVal) Utf8() string {
	switch s.dt {
	case genval.UTF8:
		return s.utf8
	case genval.Bytes:
		return string(s.bytes)
	default:
		return ""
	}
}

func (s StringVal) Bytes() []byte {
	switch s.dt {
	case genval.UTF8:
		return []byte(s.utf8)
	case genval.Bytes:
		return s.bytes
	default:
		return []byte{}
	}
}
