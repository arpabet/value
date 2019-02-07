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
)

const (

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

	mpNegFixIntPrefix 	byte = 0xe0

	defBufSize = 16
)

var (
	mpNilBin 	=  []byte { mpNil }
	mpTrueBin 	=  []byte { mpTrue }
	mpFalseBin 	=  []byte { mpFalse }
)


type messagePacker struct {
	buf 	[defBufSize]byte
	w 		io.Writer
	err     error
}

func MessagePacker(w io.Writer) *messagePacker {
	return &messagePacker{w: w}
}

func (p messagePacker) PackNil()  {
	if p.err == nil {
		_, p.err = p.w.Write(p.writeNil())
	}
}

func (p messagePacker) PackBool(val bool) {
	if p.err == nil {
		_, p.err = p.w.Write(p.writeBool(val))
	}
}

func (p messagePacker) PackLong(val int64) {
	if p.err == nil {
		_, p.err = p.w.Write(p.writeVLong(val))
	}
}

func (p messagePacker) PackDouble(val float64) {
	if p.err == nil {
		_, p.err = p.w.Write(p.writeDouble(val))
	}
}

func (p messagePacker) PackString(str string) {
	b := []byte(str)
	if p.err == nil {
		_, p.err = p.w.Write(p.writeStrHeader(len(b)))
	}
	if p.err == nil {
		_, p.err = p.w.Write(b)
	}
}

func (p messagePacker) PackBytes(b []byte) {
	if p.err == nil {
		_, p.err = p.w.Write(p.writeBinHeader(len(b)))
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
		_, p.err = p.w.Write(p.writeArrayHeader(size))
	}
}

func (p messagePacker) PackMap(size int) {
	if size < 0 {
		size = 0
	}
	if p.err == nil {
		_, p.err = p.w.Write(p.writeMapHeader(size))
	}
}

func (p messagePacker) Error() error {
	return p.err
}

func (p messagePacker) writeNil() []byte {
	return mpNilBin
}

func (p messagePacker) writeBool(val bool) []byte {
	if val {
		return mpTrueBin
	} else {
		return mpFalseBin
	}
}

func (p messagePacker) writeVLong(val int64) []byte {

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

func (p messagePacker) writeVULong(val uint64) []byte {
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

func (p messagePacker) writeDouble(val float64) []byte {
	p.buf[0] = mpFloat64
	binary.BigEndian.PutUint64(p.buf[1:9], math.Float64bits(val))
	return p.buf[:9]
}

func (p messagePacker) writeBinHeader(len int) []byte {
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

func (p messagePacker) writeStrHeader(len int) []byte {
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

func (p messagePacker) writeArrayHeader(len int) []byte {
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

func (p messagePacker) writeMapHeader(len int) []byte {
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