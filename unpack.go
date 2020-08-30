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
	"github.com/pkg/errors"
	"io"
	"strconv"
)


/**
	@author Alex Shvid
*/

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
		return doParseList(header, unpacker, parser)
	case MapHeader:
		return doParseMap(header, unpacker, parser)
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

func doParseList(header []byte, unpacker Unpacker, parser Parser) (List, error) {
	cnt := parser.ParseList(header)
	if parser.Error() != nil {
		return nil, parser.Error()
	}
	if cnt == 0 {
		return EmptyList(), nil
	}
	list := make([]Value, cnt)
	for i := 0; i < cnt; i++ {
		el, err := doParse(unpacker, parser)
		if err != nil {
			return nil, err
		}
		list[i] = el
	}
	return SolidList(list), nil
}

func doParseMap(header []byte, unpacker Unpacker, parser Parser) (Value, error) {
	cnt := parser.ParseMap(header)
	if parser.Error() != nil {
		return nil, parser.Error()
	}
	if cnt == 0 {
		return EmptyMap(), nil
	}
	var sparseListItems []ListItem
	mayBeList := false
	var sortedMapEntries []MapEntry
	sorted := true
	var prevListKey int64
	var prevMapKey string

	for i := 0; i < cnt; i++ {
		key, err := doParse(unpacker, parser)
		if err != nil {
			return nil, err
		}
		value, err := doParse(unpacker, parser)
		if err != nil {
			return nil, err
		}

		if key == nil {
			// nothing to do with this
			continue
		}

		// first element
		if i == 0 {
			if key.Kind() == NUMBER {
				// try to build sparse list
				mayBeList = true
				sparseListItems = make([]ListItem, cnt)
			} else {
				mayBeList = false
				sortedMapEntries = make([]MapEntry, cnt)
			}
		}

		if mayBeList {

			if key.Kind() == NUMBER {
				k := key.(Number).Long()
				if i > 0 && prevListKey > k {
					sorted = false
				}
				sparseListItems[i] = Item(int(k), value)
				prevListKey = k
			} else {
				// not a list
				mayBeList = false
				sortedMapEntries = make([]MapEntry, cnt)
				for j := 0; j < i; j++ {
					item := sparseListItems[i]
					sortedMapEntries[i] = Entry(strconv.Itoa(item.Key()), item.Value())
				}
				k := key.String()
				if i > 0 && prevMapKey > k {
					sorted = false
				}
				sortedMapEntries[i] = Entry(k, value)
				prevMapKey = k
			}

		} else {
			k := key.String()
			if i > 0 && prevMapKey > k {
				sorted = false
			}
			sortedMapEntries[i] = Entry(k, value)
			prevMapKey = k
		}

	}

	if mayBeList {
		return SparseList(sparseListItems, sorted), nil
	} else {
		return SortedMap(sortedMapEntries, sorted), nil
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
