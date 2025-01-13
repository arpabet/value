/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package value


import (
	"bytes"
	"encoding/base64"
	"reflect"
	"strings"
)

var UnknownPrefix = "data:application/x-msgpack-ext;"

/**
	first byte is the xtag
	then data
*/
type unknownValue []byte

func Unknown(tagAndData []byte) unknownValue {
	return unknownValue(tagAndData)
}

func (x unknownValue) Kind() Kind {
	return UNKNOWN
}

func (x unknownValue) Class() reflect.Type {
	return reflect.TypeOf((*unknownValue)(nil)).Elem()
}

func (v unknownValue) String() string {
	var out strings.Builder
	out.WriteString(UnknownPrefix)
	out.WriteString(Base64Prefix)
	out.WriteString(base64.RawStdEncoding.EncodeToString(v))
	return out.String()
}

func (x unknownValue) Tag() Ext {
	return Ext(x[0])
}

func (x unknownValue) Data() []byte {
	return x[1:]
}

func (x unknownValue) Native() []byte {
	return x
}

func (x unknownValue) Object() interface{} {
	return []byte(x)
}

func (x unknownValue) Pack(p Packer) {
	p.PackExt(x.Tag(), x.Data())
}

func (v unknownValue) PrintJSON(out *strings.Builder) {
	out.WriteRune(jsonQuote)
	out.WriteString(UnknownPrefix)
	out.WriteString(Base64Prefix)
	out.WriteString(base64.RawStdEncoding.EncodeToString(v))
	out.WriteRune(jsonQuote)
}

func (v unknownValue) MarshalJSON() ([]byte, error) {
	var out strings.Builder
	out.WriteRune(jsonQuote)
	out.WriteString(UnknownPrefix)
	out.WriteString(Base64Prefix)
	out.WriteString(base64.RawStdEncoding.EncodeToString(v))
	out.WriteRune(jsonQuote)
	return []byte(out.String()), nil
}

func (x unknownValue) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	p := MessagePacker(&buf)
	x.Pack(p)
	return buf.Bytes(), p.Error()
}

func (x unknownValue) Equal(val Value) bool {
	if val == nil || val.Kind() != UNKNOWN {
		return false
	}
	if o, ok := val.(Extension); ok {
		return bytes.Compare(x.Native(), o.Native()) == 0
	}
	return false
}
