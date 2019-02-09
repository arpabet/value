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
	"strings"
	"bytes"
	"encoding/base64"
)

var DataPrefix = "data:application/x-msgpack;"

type unknownValue struct {
	packed   []byte
}

func (v unknownValue) Equal(val Value) bool {
	if val == nil || val.Kind() != UNKNOWN {
		return false
	}
	o := val.(*unknownValue)
	return bytes.Compare(v.packed, o.packed) == 0
}

func Unknown(header, payload []byte) *unknownValue {

	size := len(header) + len(payload)
	p := make([]byte, size)

	copy(p, header)
	if payload != nil {
		copy(p[len(header):], payload)
	}

	return &unknownValue{p}
}

func (v unknownValue) Kind() Kind {
	return UNKNOWN
}

func (v unknownValue) Class() reflect.Type {
	return reflect.TypeOf((*unknownValue)(nil)).Elem()
}

func (v unknownValue) String() string {
	var out strings.Builder
	out.WriteString(DataPrefix)
	out.WriteString(Base64Prefix)
	out.WriteString(base64.RawStdEncoding.EncodeToString(v.packed))
	return out.String()
}

func (v unknownValue) Packed() []byte {
	return v.packed
}

func (v unknownValue) Pack(p Packer) {
	p.PackUnknown(v.packed)
}

func (v unknownValue) PrintJSON(out *strings.Builder) {
	out.WriteRune(jsonQuote)
	out.WriteString(DataPrefix)
	out.WriteString(Base64Prefix)
	out.WriteString(base64.RawStdEncoding.EncodeToString(v.packed))
	out.WriteRune(jsonQuote)
}

func (v unknownValue) MarshalJSON() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v unknownValue) MarshalBinary() ([]byte, error) {
	return v.packed, nil
}
