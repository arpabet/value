/**
    Copyright (c) 2020-2022 Arpabet, Inc.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in
	all copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
	THE SOFTWARE.
*/

package value

/**
Base interface for the packing values
*/

type Ext byte

const (
	UnknownExt Ext = iota

	BigIntExt
	DecimalExt

	MaxExt
)

type Packer interface {
	PackNil()

	PackBool(bool)

	PackLong(int64)

	PackDouble(float64)

	PackStr(string)

	PackBin([]byte)

	PackExt(xtag Ext, data []byte)

	PackList(int)

	PackMap(int)

	PackRaw([]byte)

	Error() error
}

/**

Interface for write values

*/

type Writer interface {
	WriteNil() []byte

	WriteBool(val bool) []byte

	WriteLong(val int64) []byte

	WriteDouble(val float64) []byte

	WriteBinHeader(len int) []byte

	WriteStrHeader(len int) []byte

	WriteExtHeader(len int, xtag byte) []byte

	WriteArrayHeader(len int) []byte

	WriteMapHeader(len int) []byte
}

/**
Base interface for the unpacking values

*/

type Format int

const (
	EOF Format = iota
	UnexpectedEOF
	NilToken
	BoolToken
	LongToken
	DoubleToken
	FixExtToken
	BinHeader
	StrHeader
	ListHeader
	MapHeader
	ExtHeader
)

type Unpacker interface {
	Next() (Format, []byte)

	Read(int) ([]byte, error)
}

/**
	Parse value from Slice, the slice size must be enough to parse primitive value or header, slice always stars from code

    return - number of bytes read, value

	return 0 read bytes on error
*/

type Parser interface {
	ParseBool([]byte) bool

	ParseLong([]byte) int64

	ParseDouble([]byte) float64

	ParseBin([]byte) int

	ParseStr([]byte) int

	ParseList([]byte) int

	ParseMap([]byte) int

	ParseExt([]byte) (int, []byte)

	Error() error
}
