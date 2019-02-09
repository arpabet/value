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

package genval

import (
	"bytes"
	"encoding/hex"
	"strings"
	"io"
	"github.com/pkg/errors"
)

func Pack(val Value) []byte {

	buf := bytes.Buffer{}
	p := MessagePacker(&buf)
	val.Pack(p)

	return buf.Bytes()
}

func Unpack(buf []byte, copy bool) (Value, error) {
	unpacker := MessageUnpacker(buf, copy)
	parser := MessageParser()
	return Parse(unpacker, parser)
}

func Read(r io.Reader) (Value, error) {
	unpacker := MessageReader(r)
	parser := MessageParser()
	return Parse(unpacker, parser)
}

func Hex(val Value) string {
	return hex.EncodeToString(Pack(val))
}

func Json(val Value) string {
	var out strings.Builder
	val.PrintJSON(&out)
	return out.String()
}

func Parse(unpacker Unpacker, parser Parser) (Value, error) {
	return doParse(unpacker, parser)
}

func Stream(r io.Reader, out chan<- Value) error {

	defer close(out)

	unpacker := MessageReader(r)
	parser := MessageParser()

	for {

		value, err := doParse(unpacker, parser)
		if err != nil {
			return err
		}

		out <- value
	}

	return nil
}

func doParse(unpacker Unpacker, parser Parser) (Value, error) {

	format, header := unpacker.Next()

	switch format {
	case EOF:
		return nil, nil
	case UnexpectedEOF:
		return nil, io.ErrUnexpectedEOF
	case NilToken:
		return nil, nil
	case BoolToken:
		return Boolean(parser.ParseBool(header)), parser.Error()
	case LongToken:
		return Long(parser.ParseLong(header)), parser.Error()
	case DoubleToken:
		return Double(parser.ParseDouble(header)), parser.Error()
	case FixExtToken:
		return Unknown(header, nil), nil
	case BinHeader:
		size := parser.ParseBin(header)
		if parser.Error() != nil {
			return nil, parser.Error()
		}
		raw, err := unpacker.Read(size)
		if err != nil {
			return nil, err
		}
		return Raw(raw, false), nil
	case StrHeader:
		len := parser.ParseStr(header)
		if parser.Error() != nil {
			return nil, parser.Error()
		}
		str, err := unpacker.Read(len)
		if err != nil {
			return nil, err
		}
		return Utf8(string(str)), nil
	case ListHeader:
		cnt := parser.ParseList(header)
		if parser.Error() != nil {
			return nil, parser.Error()
		}
		list := List()
		for i:=0; i<cnt; i++ {
			el, err := doParse(unpacker, parser)
			if err != nil {
				return nil, err
			}
			if el != nil {
				// nil elements are not supported
				list.Insert(el)
			}
		}
		return list, nil
	case MapHeader:
		cnt := parser.ParseMap(header)
		if parser.Error() != nil {
			return nil, parser.Error()
		}
		m := Map()
		for i:=0; i<cnt; i++ {
			key, err := doParse(unpacker, parser)
			if err != nil {
				return nil, err
			}
			value, err := doParse(unpacker, parser)
			if err != nil {
				return nil, err
			}
			if key != nil && value != nil {
				// only non-null key-value pairs are supported
				if key.Kind() == NUMBER {
					m.PutAt(int(key.(Number).Long()), value)
				} else {
					m.Put(key.String(), value)
				}
			}
		}
		return m, nil
	case ExtHeader:
		size := parser.ParseExt(header)
		if parser.Error() != nil {
			return nil, parser.Error()
		}
		size += 1  // add tag
		ext, err := unpacker.Read(size)
		if err != nil {
			return nil, err
		}
		return Unknown(header, ext), nil
	default:
		return nil, errors.Errorf("parse: invalid format %v", format)
	}

}