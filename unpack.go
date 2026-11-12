/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */


package value

import (
	"golang.org/x/xerrors"
	"io"
	"strconv"
)


func doParse(unpacker Unpacker, parser Parser, depth int) (Value, error) {

	if err := checkDepth(depth); err != nil {
		return nil, err
	}

	format, header := unpacker.Next()

	switch format {
	case EOF:
		return nil, io.EOF
	case UnexpectedEOF:
		return nil, io.ErrUnexpectedEOF
	case NilToken:
		return Null, nil
	case BoolToken:
		return Boolean(parser.ParseBool(header)), parser.Error()
	case LongToken:
		n := parser.ParseLong(header)
		if n < 0 {
			// MessagePack's uint 64 holds integers a long does not, and a
			// long read from one is negative. A canonical writer packs such
			// an integer as a big integer; one from another writer reads as
			// the same number.
			if up, ok := parser.(unsignedParser); ok {
				if u, beyond := up.parseUnsigned(header); beyond {
					return Unsigned(u), parser.Error()
				}
			}
		}
		return Long(n), parser.Error()
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
		if err := checkByteLen(size); err != nil {
			return nil, err
		}
		raw, err := unpacker.Read(size)
		if err != nil {
			return nil, err
		}
		return Raw(raw, false), nil
	case StrHeader:
		size := parser.ParseStr(header)
		if parser.Error() != nil {
			return nil, parser.Error()
		}
		if err := checkByteLen(size); err != nil {
			return nil, err
		}
		str, err := unpacker.Read(size)
		if err != nil {
			return nil, err
		}
		return Utf8(string(str)), nil
	case ListHeader:
		return doParseList(header, unpacker, parser, depth)
	case MapHeader:
		return doParseMap(header, unpacker, parser, depth)
	case ExtHeader:
		n, _ := parser.ParseExt(header)
		if parser.Error() != nil {
			return nil, parser.Error()
		}
		if err := checkByteLen(n); err != nil {
			return nil, err
		}
		tagAndData, err := unpacker.Read(n+1)
		if err != nil {
			return nil, err
		}
		return doParseExt(tagAndData)
	default:
		return nil, xerrors.Errorf("parse: invalid format %v", format)
	}

}

// unsignedParser is a Parser whose integer tokens can hold a number above the
// range of a long.
type unsignedParser interface {
	parseUnsigned(header []byte) (val uint64, beyond bool)
}

func doParseList(header []byte, unpacker Unpacker, parser Parser, depth int) (List, error) {
	cnt := parser.ParseList(header)
	if parser.Error() != nil {
		return nil, parser.Error()
	}
	if err := checkCollectionLen(cnt); err != nil {
		return nil, err
	}
	if cnt == 0 {
		return EmptyImmutableList(), nil
	}
	// Grow incrementally from a bounded starting capacity so a header that
	// claims a huge length but provides little data cannot force a large
	// allocation: the loop errors out at the first missing element.
	list := make([]Value, 0, initialSliceCap(cnt))
	for i := 0; i < cnt; i++ {
		el, err := doParse(unpacker, parser, depth+1)
		if err != nil {
			return nil, err
		}
		list = append(list, el)
	}
	return ImmutableList(list), nil
}

func doParseMap(header []byte, unpacker Unpacker, parser Parser, depth int) (Value, error) {
	cnt := parser.ParseMap(header)
	if parser.Error() != nil {
		return nil, parser.Error()
	}
	if err := checkCollectionLen(cnt); err != nil {
		return nil, err
	}
	if cnt == 0 {
		return EmptyImmutableMap(), nil
	}
	initCap := initialSliceCap(cnt)
	var sparseListItems []ListItem
	mayBeList := false
	var sortedMapEntries []MapEntry
	sorted := true
	var prevListKey int64
	var prevMapKey string

	for i := 0; i < cnt; i++ {
		key, err := doParse(unpacker, parser, depth+1)
		if err != nil {
			return nil, err
		}
		value, err := doParse(unpacker, parser, depth+1)
		if err != nil {
			return nil, err
		}

		if key == nil {
			// nothing to do with this
			continue
		}

		// first element
		if i == 0 {
			if isListIndexKey(key) {
				// try to build sparse list
				mayBeList = true
				sparseListItems = make([]ListItem, 0, initCap)
			} else {
				mayBeList = false
				sortedMapEntries = make([]MapEntry, 0, initCap)
			}
		}

		if mayBeList {

			if isListIndexKey(key) {
				k := key.(Number).Long()
				if i > 0 && prevListKey > k {
					sorted = false
				}
				sparseListItems = append(sparseListItems, ImmutableItem(int(k), value))
				prevListKey = k
			} else {
				// not a list: fall back to a string-keyed map, back-filling the
				// entries already decoded as list items. The mix of stringified-int
				// and string keys has no guaranteed order, so force a re-sort.
				mayBeList = false
				sorted = false
				sortedMapEntries = make([]MapEntry, 0, initCap)
				for _, item := range sparseListItems {
					sortedMapEntries = append(sortedMapEntries, ImmutableEntry(strconv.Itoa(item.Key()), item.Value()))
				}
				k := key.String()
				sortedMapEntries = append(sortedMapEntries, ImmutableEntry(k, value))
				prevMapKey = k
			}

		} else {
			k := key.String()
			if i > 0 && prevMapKey > k {
				sorted = false
			}
			sortedMapEntries = append(sortedMapEntries, ImmutableEntry(k, value))
			prevMapKey = k
		}

	}

	if mayBeList {
		return SparseList(sparseListItems, sorted), nil
	} else {
		return ImmutableMap(sortedMapEntries, sorted), nil
	}

}

// isListIndexKey reports whether a decoded map key should be treated as a sparse
// list index. Only fixed integer (LONG) keys qualify: the library only ever
// encodes sparse-list indices as LONG, and converting a hostile decimal or big
// integer key to int64 could be very expensive. The index must also be a sane,
// non-negative value within MaxParseCollectionLen; anything else stays a
// string-keyed map entry, so sparseList.Len()/Values() can never materialize a
// huge or negative slice from hostile input.
func isListIndexKey(key Value) bool {
	if key.Kind() != NUMBER {
		return false
	}
	n := key.(Number)
	if n.Type() != LONG {
		return false
	}
	k := n.Long()
	return k >= 0 && (MaxParseCollectionLen <= 0 || k <= int64(MaxParseCollectionLen))
}

func doParseExt(tagAndData []byte) (Value, error) {
	if len(tagAndData) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	xtag := Ext(tagAndData[0])
	ext := tagAndData[1:]
	switch xtag {

	case BigIntExt:
		if LegacyBigNumbers {
			v, err := UnpackLegacyBigInt(ext)
			if err != nil {
				return nil, err
			}
			return BigInt(v), nil
		}
		v, err := UnpackBigInt(ext)
		if err != nil {
			return nil, err
		}
		return BigInt(v), nil
	case DecimalExt:
		if LegacyBigNumbers {
			v, err := UnpackLegacyDecimal(ext)
			if err != nil {
				return nil, err
			}
			return Decimal(v), nil
		}
		v, err := UnpackDecimal(ext)
		if err != nil {
			return nil, err
		}
		return Decimal(v), nil
	case TimestampExt:
		return UnpackTimestamp(ext)

	}
	return Unknown(tagAndData), nil
}
