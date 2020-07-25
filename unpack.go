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
	"encoding/hex"
	"strings"
	"io"
	"github.com/pkg/errors"
)

func Pack(val Value) ([]byte, error) {
	buf := bytes.Buffer{}
	p := MessagePacker(&buf)
	val.Pack(p)
	return buf.Bytes(), p.Error()
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

func Write(w io.Writer, val Value) error {
	p := MessagePacker(w)
	val.Pack(p)
	return p.Error()
}

func Hex(val Value) string {
	mp, _ := Pack(val)
	return hex.EncodeToString(mp)
}

func Json(val Value) string {
	var out strings.Builder
	val.PrintJSON(&out)
	return out.String()
}

func Parse(unpacker Unpacker, parser Parser) (Value, error) {
	return doParse(unpacker, parser)
}

func WriteStream(w io.Writer, valueC <-chan Value) error {

	p := MessagePacker(w)

	for p.Error() == nil {
		val, ok := <- valueC

		if !ok {
			break
		}

		val.Pack(p)

	}

	return p.Error()
}

func ReadStream(r io.Reader, out chan<- Value) error {

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
		return nil, io.EOF
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
		_, tagAndData := parser.ParseExt(header)
		return doParseExt(tagAndData)
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
		n, _ := parser.ParseExt(header)
		if parser.Error() != nil {
			return nil, parser.Error()
		}
		tagAndData, err := unpacker.Read(n+1)
		if err != nil {
			return nil, err
		}
		return doParseExt(tagAndData)
	default:
		return nil, errors.Errorf("parse: invalid format %v", format)
	}

}

func doParseExt(tagAndData []byte) (Value, error) {
	xtag := Ext(tagAndData[0])
	ext := tagAndData[1:]
	switch xtag {

	case BigIntExt:
		v, err := UnpackBigInt(ext)
		return BigInt(v), err
	case DecimalExt:
		v, err := UnpackDecimal(ext)
		return Decimal(v), err

	}
	return Unknown(tagAndData), nil
}
