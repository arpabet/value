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


type MessagePacker struct {
	buf 	[defBufSize]byte
	w 		io.Writer
	err     error
}

func NewMessagePacker(w io.Writer) *MessagePacker {
	return &MessagePacker{w: w}
}

func (p MessagePacker) PackNil()  {
	if p.err == nil {
		_, p.err = p.w.Write(mpNilBin)
	}
}

func (p MessagePacker) PackBool(val bool) {
	if p.err == nil {
		if val {
			_, p.err = p.w.Write(mpTrueBin)
		} else {
			_, p.err = p.w.Write(mpFalseBin)
		}
	}
}

func (p MessagePacker) PackLong(val int64) {
	if p.err == nil {
		_, p.err = p.w.Write(p.writeVLong(val))
	}
}

func (p MessagePacker) PackDouble(val float64) {
	if p.err == nil {
		p.buf[0] = mpFloat64
		binary.BigEndian.PutUint64(p.buf[1:9], math.Float64bits(val))
		_, p.err = p.w.Write(p.buf[:9])
	}
}

func (p MessagePacker) PackString(str string) {
	b := []byte(str)
	if p.err == nil {
		_, p.err = p.w.Write(p.writeStrHeader(len(b)))
	}
	if p.err == nil {
		_, p.err = p.w.Write(b)
	}
}

func (p MessagePacker) PackBytes(b []byte) {
	if p.err == nil {
		_, p.err = p.w.Write(p.writeBinHeader(len(b)))
	}
	if p.err == nil {
		_, p.err = p.w.Write(b)
	}
}

func (p MessagePacker) Error() error {
	return p.err
}

func (p MessagePacker) writeVLong(val int64) []byte {

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

func (p MessagePacker) writeVULong(val uint64) []byte {
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

func (p MessagePacker) writeFixedULong(val uint64) []byte {
	b := p.buf[:8]
	binary.BigEndian.PutUint64(b, val)
	return b
}

func (p MessagePacker) writeBinHeader(len int) []byte {
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

func (p MessagePacker) writeStrHeader(len int) []byte {
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
