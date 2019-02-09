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
	"io"
	"math"
	"encoding/binary"
	"github.com/pkg/errors"
)

const (

	mpPosFixIntMin 		byte = 0x00
	mpPosFixIntMax      byte = 0x7f
	mpFixMapMin         byte = 0x80
	mpFixMapMax         byte = 0x8f
	mpFixArrayMin       byte = 0x90
	mpFixArrayMax       byte = 0x9f
	mpFixStrMin         byte = 0xa0
	mpFixStrMax         byte = 0xbf

	mpPosFixIntMask 	byte = 0x80
	mpFixMapPrefix  	byte = 0x80
	mpFixArrayPrefix  	byte = 0x90
	mpFixStrPrefix    	byte = 0xa0

	mpNil          		byte = 0xc0
	mpNeverUsed    		byte = 0xc1
	mpFalse        		byte = 0xc2
	mpTrue         		byte = 0xc3

	mpBin8     			byte = 0xc4
	mpBin16    			byte = 0xc5
	mpBin32    			byte = 0xc6
	mpExt8     			byte = 0xc7
	mpExt16    			byte = 0xc8
	mpExt32    			byte = 0xc9

	mpFloat32   		byte = 0xca
	mpFloat64   		byte = 0xcb

	mpUint8        		byte = 0xcc
	mpUint16       		byte = 0xcd
	mpUint32       		byte = 0xce
	mpUint64       		byte = 0xcf

	mpInt8         		byte = 0xd0
	mpInt16        		byte = 0xd1
	mpInt32        		byte = 0xd2
	mpInt64        		byte = 0xd3

	mpFixExt1  			byte = 0xd4
	mpFixExt2  			byte = 0xd5
	mpFixExt4  			byte = 0xd6
	mpFixExt8  			byte = 0xd7
	mpFixExt16 			byte = 0xd8

	mpStr8  			byte = 0xd9
	mpStr16 			byte = 0xda
	mpStr32 			byte = 0xdb

	mpArray16 			byte = 0xdc
	mpArray32 			byte = 0xdd

	mpMap16 			byte = 0xde
	mpMap32 			byte = 0xdf

	mpNegFixIntMin 		byte = 0xe0
	mpNegFixIntMax 		byte = 0xff

	mpNegFixIntPrefix 	byte = 0xe0

	defWriteBufSize 	= 16
	defReadBufSize 		= 24

	mpCodeMin 			= mpNil
	mpCodeMax 			= mpMap32
)


var (

	mpCodeFormat = []Format {

		NilToken, 	// mpNil        0xc0
		NilToken,     // mpNeverUsed  0xc1
		BoolToken, 	// mpFalse      0xc2
		BoolToken, 	// mpTrue       0xc3

		BinHeader,  // mpBin8		0xc4
		BinHeader,  // mpBin16		0xc5
		BinHeader,  // mpBin32		0xc6
		ExtHeader,  // mpExt8		0xc7
		ExtHeader,  // mpExt16		0xc8
		ExtHeader,  // mpExt32		0xc9

		DoubleToken,  // mpFloat32	0xca
		DoubleToken,  // mpFloat64	0xcb

		LongToken,  // mpUint8		0xcc
		LongToken,  // mpUint16		0xcd
		LongToken,  // mpUint32		0xce
		LongToken,  // mpUint64  	0xcf

		LongToken,  // mpInt8 		0xd0
		LongToken,  // mpInt16	 	0xd1
		LongToken,  // mpInt32   	0xd2
		LongToken,  // mpInt64  	0xd3

		FixExtToken,  // mpFixExt1  	0xd4
		FixExtToken,  // mpFixExt2  	0xd5
		FixExtToken,  // mpFixExt4  	0xd6
		FixExtToken,  // mpFixExt8  	0xd7
		FixExtToken, // mpFixExt16 	0xd8

		StrHeader,  // mpStr8  		0xd9
		StrHeader,  // mpStr16 	 	0xda
		StrHeader,  // mpStr32 		0xdb

		ListHeader,  // mpArray16 	0xdc
		ListHeader,  // mpArray32 	0xdd

		MapHeader,  // mpMap16 	 	0xde
		MapHeader,  // mpMap32 		0xdf

	}

	mpCodeSize = []int {

		0, 	// mpNil        0xc0
		0, 	// mpNeverUsed  0xc1
		0, 	// mpFalse      0xc2
		0, 	// mpTrue       0xc3

		1,  // mpBin8		0xc4
		2,  // mpBin16		0xc5
		4,  // mpBin32		0xc6
		1,  // mpExt8		0xc7
		2,  // mpExt16		0xc8
		4,  // mpExt32		0xc9

		4,  // mpFloat32	0xca
		8,  // mpFloat64	0xcb

		1,  // mpUint8		0xcc
		2,  // mpUint16		0xcd
		4,  // mpUint32		0xce
		8,  // mpUint64  	0xcf

		1,  // mpInt8 		0xd0
		2,  // mpInt16	 	0xd1
		4,  // mpInt32   	0xd2
		8,  // mpInt64  	0xd3

		1,  // mpFixExt1  	0xd4
		2,  // mpFixExt2  	0xd5
		4,  // mpFixExt4  	0xd6
		8,  // mpFixExt8  	0xd7
		16, // mpFixExt16 	0xd8

		1,  // mpStr8  		0xd9
		2,  // mpStr16 	 	0xda
		4,  // mpStr32 		0xdb

		2,  // mpArray16 	0xdc
		4,  // mpArray32 	0xdd

		2,  // mpMap16 	 	0xde
		4,  // mpMap32 		0xdf

	}

	mpNilBin 	=  []byte { mpNil }
	mpTrueBin 	=  []byte { mpTrue }
	mpFalseBin 	=  []byte { mpFalse }
)

type messagePacker struct {
	m   messageWriter
	w   io.Writer
	err error
}

func MessagePacker(w io.Writer) *messagePacker {
	return &messagePacker{w: w}
}

func (p messagePacker) PackNil()  {
	if p.err == nil {
		_, p.err = p.w.Write(p.m.WriteNil())
	}
}

func (p messagePacker) PackBool(val bool) {
	if p.err == nil {
		_, p.err = p.w.Write(p.m.WriteBool(val))
	}
}

func (p messagePacker) PackLong(val int64) {
	if p.err == nil {
		_, p.err = p.w.Write(p.m.WriteLong(val))
	}
}

func (p messagePacker) PackDouble(val float64) {
	if p.err == nil {
		_, p.err = p.w.Write(p.m.WriteDouble(val))
	}
}

func (p messagePacker) PackStr(str string) {
	b := []byte(str)
	if p.err == nil {
		_, p.err = p.w.Write(p.m.WriteStrHeader(len(b)))
	}
	if p.err == nil {
		_, p.err = p.w.Write(b)
	}
}

func (p messagePacker) PackBin(b []byte) {
	if p.err == nil {
		_, p.err = p.w.Write(p.m.WriteBinHeader(len(b)))
	}
	if p.err == nil {
		_, p.err = p.w.Write(b)
	}
}

func (p messagePacker) PackList(size int) {
	if size < 0 {
		size = 0
	}
	if p.err == nil {
		_, p.err = p.w.Write(p.m.WriteArrayHeader(size))
	}
}

func (p messagePacker) PackMap(size int) {
	if size < 0 {
		size = 0
	}
	if p.err == nil {
		_, p.err = p.w.Write(p.m.WriteMapHeader(size))
	}
}

func (p messagePacker) PackUnknown(b []byte) {
	if p.err == nil {
		_, p.err = p.w.Write(b)
	}
}

func (p messagePacker) Error() error {
	return p.err
}

type messageWriter struct {
	buf 	[defWriteBufSize]byte
}

func (p messageWriter) WriteNil() []byte {
	return mpNilBin
}

func (p messageWriter) WriteBool(val bool) []byte {
	if val {
		return mpTrueBin
	} else {
		return mpFalseBin
	}
}

func (p messageWriter) WriteLong(val int64) []byte {

	switch {
		case val >= 0:
			return p.writeVULong(uint64(val))
		case val >= -32:
			p.buf[0] = byte(val)
			return p.buf[:1]
		case val >= math.MinInt8:
			p.buf[0] = mpInt8
			p.buf[1] = byte(val)
			return p.buf[:2]
		case val >= math.MinInt16:
			p.buf[0] = mpInt16
			binary.BigEndian.PutUint16(p.buf[1:3], uint16(val))
			return p.buf[:3]
		case val >= math.MinInt32:
			p.buf[0] = mpInt32
			binary.BigEndian.PutUint32(p.buf[1:5], uint32(val))
			return p.buf[:5]
		default:
			p.buf[0] = mpInt64
			binary.BigEndian.PutUint64(p.buf[1:9], uint64(val))
			return p.buf[:9]
	}

}

func (p messageWriter) writeVULong(val uint64) []byte {
	switch {
	case val <= math.MaxInt8:
		p.buf[0] = byte(val)
		return p.buf[:1]
	case val <= math.MaxUint8:
		p.buf[0] = mpUint8
		p.buf[1] = byte(val)
		return p.buf[:2]
	case val <= math.MaxUint16:
		p.buf[0] = mpUint16
		binary.BigEndian.PutUint16(p.buf[1:3], uint16(val))
		return p.buf[:3]
	case val <= math.MaxUint32:
		p.buf[0] = mpUint32
		binary.BigEndian.PutUint32(p.buf[1:5], uint32(val))
		return p.buf[:5]
	default:
		p.buf[0] = mpUint64
		binary.BigEndian.PutUint64(p.buf[1:9], val)
		return p.buf[:9]
	}
}

func (p messageWriter) WriteDouble(val float64) []byte {
	p.buf[0] = mpFloat64
	binary.BigEndian.PutUint64(p.buf[1:9], math.Float64bits(val))
	return p.buf[:9]
}

func (p messageWriter) WriteBinHeader(len int) []byte {
	switch {
	case len <= math.MaxUint8:
		p.buf[0] = mpBin8
		p.buf[1] = byte(len)
		return p.buf[:2]
	case len <= math.MaxUint16:
		p.buf[0] = mpBin16
		binary.BigEndian.PutUint16(p.buf[1:3], uint16(len))
		return p.buf[:3]
	default:
		p.buf[0] = mpBin32
		binary.BigEndian.PutUint32(p.buf[1:5], uint32(len))
		return p.buf[:5]
	}
}

func (p messageWriter) WriteStrHeader(len int) []byte {
	switch {
	case len < 32:
		p.buf[0] = mpFixStrPrefix | byte(len)
		return p.buf[:1]
	case len <= math.MaxUint8:
		p.buf[0] = mpStr8
		p.buf[1] = byte(len)
		return p.buf[:2]
	case len <= math.MaxUint16:
		p.buf[0] = mpStr16
		binary.BigEndian.PutUint16(p.buf[1:3], uint16(len))
		return p.buf[:3]
	default:
		p.buf[0] = mpStr32
		binary.BigEndian.PutUint32(p.buf[1:5], uint32(len))
		return p.buf[:5]
	}
}

func (p messageWriter) WriteArrayHeader(len int) []byte {
	switch {
	case len < 16:
		p.buf[0] = mpFixArrayPrefix | byte(len)
		return p.buf[:1]
	case len <= math.MaxUint16:
		p.buf[0] = mpArray16
		binary.BigEndian.PutUint16(p.buf[1:3], uint16(len))
		return p.buf[:3]
	default:
		p.buf[0] = mpArray16
		binary.BigEndian.PutUint32(p.buf[1:5], uint32(len))
		return p.buf[:5]
	}
}

func (p messageWriter) WriteMapHeader(len int) []byte {
	switch {
	case len < 16:
		p.buf[0] = mpFixMapPrefix | byte(len)
		return p.buf[:1]
	case len <= math.MaxUint16:
		p.buf[0] = mpMap16
		binary.BigEndian.PutUint16(p.buf[1:3], uint16(len))
		return p.buf[:3]
	default:
		p.buf[0] = mpMap32
		binary.BigEndian.PutUint32(p.buf[1:5], uint32(len))
		return p.buf[:5]
	}
}

type messageParser struct {
	err  error
}

func MessageParser() *messageParser {
	return &messageParser{}
}

func (r *messageParser) ParseBool(b []byte) bool {

	code := b[0]

	switch code {
	case mpTrue:
		return true
	case mpFalse:
		return false
	default:
		r.err = errors.Errorf("bool: invalid code %v", code)
		return false
	}
}

func (r *messageParser) ParseLong(b []byte) int64 {

	code := b[0]

	switch code {
	case mpUint8:
		return int64(b[1])
	case mpUint16:
		return int64(binary.BigEndian.Uint16(b[1:]))
	case mpUint32:
		return int64(binary.BigEndian.Uint32(b[1:]))
	case mpUint64:
		return int64(binary.BigEndian.Uint64(b[1:]))
	case mpInt8:
		return int64(int8(b[1]))
	case mpInt16:
		return int64(int16(binary.BigEndian.Uint16(b[1:])))
	case mpInt32:
		return int64(int32(binary.BigEndian.Uint32(b[1:])))
	case mpInt64:
		return int64(int64(binary.BigEndian.Uint64(b[1:])))
	}

	switch {
	case code >= mpPosFixIntMin && code <= mpPosFixIntMax:
		return int64(code)
	case code >= mpNegFixIntMin && code <= mpNegFixIntMax:
		return int64(int8(code))
	default:
		r.err = errors.Errorf("long: invalid code %v", code)
		return 0
	}
}

func (r *messageParser) ParseDouble(b []byte) float64 {

	code := b[0]

	switch code {
	case mpFloat32:
		val32 := binary.BigEndian.Uint32(b[1:])
		return float64(math.Float32frombits(val32))
	case mpFloat64:
		val64 := binary.BigEndian.Uint64(b[1:])
		return math.Float64frombits(val64)
	default:
		r.err = errors.Errorf("double: invalid code %v", code)
		return 0
	}
}

func (r *messageParser) ParseBin(b []byte) int {

	code := b[0]

	switch code {
	case mpBin8:
		return int(b[1])
	case mpBin16:
		return int(binary.BigEndian.Uint16(b[1:]))
	case mpBin32:
		return int(binary.BigEndian.Uint32(b[1:]))
	default:
		r.err = errors.Errorf("bin: invalid code %v", code)
		return 0
	}
}

func (r *messageParser) ParseStr(b []byte) int {

	code := b[0]

	if code >= mpFixStrMin && code <= mpFixStrMax {
		return int(code - mpFixStrMin)
	}

	switch code {
	case mpStr8:
		return int(b[1])
	case mpStr16:
		return int(binary.BigEndian.Uint16(b[1:]))
	case mpStr32:
		return int(binary.BigEndian.Uint32(b[1:]))
	default:
		r.err = errors.Errorf("str: invalid code %v", code)
		return 0
	}
}

func (r *messageParser) ParseList(b []byte) int {

	code := b[0]

	if code >= mpFixArrayMin && code <= mpFixArrayMax {
		return int(code - mpFixArrayMin)
	}

	switch code {
	case mpArray16:
		return int(binary.BigEndian.Uint16(b[1:]))
	case mpArray32:
		return int(binary.BigEndian.Uint32(b[1:]))
	default:
		r.err = errors.Errorf("list: invalid code %v", code)
		return 0
	}

}

func (r *messageParser) ParseMap(b []byte) int {

	code := b[0]

	if code >= mpFixMapMin && code <= mpFixMapMax {
		return int(code - mpFixMapMin)
	}

	switch code {
	case mpMap16:
		return int(binary.BigEndian.Uint16(b[1:]))
	case mpMap32:
		return int(binary.BigEndian.Uint32(b[1:]))
	default:
		r.err = errors.Errorf("map: invalid code %v", code)
		return 0
	}
}

func (r *messageParser) ParseExt(b []byte) int {

	code := b[0]

	switch code {
	case mpExt8:
		return int(b[1])
	case mpExt16:
		return int(binary.BigEndian.Uint16(b[1:]))
	case mpExt32:
		return int(binary.BigEndian.Uint32(b[1:]))
	default:
		r.err = errors.Errorf("ext: invalid code %v", code)
		return 0
	}
}

func (r messageParser) Error() error {
	return r.err
}

type messageBufUnpacker struct {
	buf  []byte
	off  int
	copy bool
}

func MessageUnpacker(buf []byte, copy bool) *messageBufUnpacker {
	return &messageBufUnpacker{buf: buf, copy: copy}
}

func (p messageBufUnpacker) remaining() int {
	return len(p.buf) - p.off
}

func (p *messageBufUnpacker) Next() (Format, []byte) {
	remaining := p.remaining()
	if remaining <= 0 {
		return EOF, nil
	}
	b := p.buf[p.off:]
	format, len := nextFormat(b[0])
	n := 1 + len
	if n > remaining {
		return UnexpectedEOF, nil
	} else {
		p.off += n
		return format, b[:n]
	}
}

func (p *messageBufUnpacker) Read(n int) ([]byte, error) {

	if p.remaining() < n {
		return nil, io.ErrUnexpectedEOF
	}

	b := p.buf[p.off:]
	p.off += n

	if p.copy {
		c := make([]byte, n)
		copy(c, b)
		return c, nil
	} else {
		return b[:n], nil
	}
}

type messageIOUnpacker struct {
	buf 	[defReadBufSize]byte
	r       io.Reader
	br      io.ByteReader  // can be null
	s		io.Seeker      // can be null
}

func MessageReader(r io.Reader) *messageIOUnpacker {
	return &messageIOUnpacker{r: r, br: r.(io.ByteReader), s: r.(io.Seeker)}
}

func (p *messageIOUnpacker) Next() (Format, []byte) {

	if p.br != nil {
		code, err := p.br.ReadByte()
		if err != nil {
			return EOF, nil
		}
		p.buf[0] = code
	}  else {
		n, _ := p.r.Read(p.buf[:1])
		if n == 0 {
			return EOF, nil
		}
	}

	format, len := nextFormat(p.buf[0])
	n := 1 + len

	m, _ := p.r.Read(p.buf[1:n])
	if m != len {
		return UnexpectedEOF, nil
	}

	return format, p.buf[0:n]
}

func (p *messageIOUnpacker) Read(n int) (b []byte, err error) {
	b = make([]byte, n)
	m, err := p.r.Read(b)
	if m != n {
		if err == nil {
			err = io.ErrUnexpectedEOF
		}
	}
	return b, err
}

func nextFormat(code byte) (Format, int) {

	switch {
		case code <= mpPosFixIntMax:
			return LongToken, 0

		case code <= mpFixMapMax:
			return MapHeader, 0

		case code <= mpFixArrayMax:
			return ListHeader, 0

		case code <= mpFixStrMax:
			return StrHeader, 0

		case code <= mpCodeMax:
			i := int(code - mpCodeMin)
			return mpCodeFormat[i], mpCodeSize[i]

		default:
			return LongToken, 0
	}

}