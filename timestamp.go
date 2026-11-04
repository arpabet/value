/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"strconv"
	"strings"
	"time"

	"golang.org/x/xerrors"
)

/*
A timestamp is an instant: seconds since the Unix epoch plus nanoseconds
within that second. It carries no location, so two time.Time values that
denote the same instant in different zones are the same timestamp and pack to
the same bytes.

The canonical encoding is MessagePack's timestamp extension (type -1) in its
smallest form: 32-bit (fixext 4) when the seconds fit an unsigned 32-bit
integer and there are no nanoseconds, 64-bit (fixext 8, 30 bits of
nanoseconds above 34 bits of seconds) when the seconds fit 34 unsigned bits,
and 96-bit (ext 8 of length 12, 32-bit nanoseconds followed by 64-bit signed
seconds) otherwise. Decoding accepts all three forms whatever their content,
and rejects nanoseconds outside [0, 1e9).
*/

const (
	nanosPerSecond = int64(time.Second)
	maxSeconds32   = int64(1)<<32 - 1
	maxSeconds34   = int64(1)<<34 - 1
)

type timestampValue struct {
	sec  int64
	nsec int32
}

var timestampValueClass = reflect.TypeOf((*timestampValue)(nil)).Elem()

// Time returns the timestamp of the instant t denotes. The location of t is
// not part of the value.
func Time(t time.Time) Timestamp {
	return timestampValue{sec: t.Unix(), nsec: int32(t.Nanosecond())}
}

// UnixTime returns the timestamp sec seconds and nsec nanoseconds after the
// Unix epoch. nsec may lie outside [0, 1e9); it is normalized into that range
// as time.Unix does.
func UnixTime(sec int64, nsec int64) Timestamp {
	if nsec < 0 || nsec >= nanosPerSecond {
		n := nsec / nanosPerSecond
		sec += n
		nsec -= n * nanosPerSecond
		if nsec < 0 {
			nsec += nanosPerSecond
			sec--
		}
	}
	return timestampValue{sec: sec, nsec: int32(nsec)}
}

// ParseTimestamp parses an RFC 3339 text such as "2026-10-30T07:28:39.075Z",
// the form String and the JSON output use.
func ParseTimestamp(str string) (Timestamp, error) {
	t, err := time.Parse(time.RFC3339Nano, str)
	if err != nil {
		return nil, xerrors.Errorf("value: parse timestamp %q: %w", str, err)
	}
	return Time(t), nil
}

// UnpackTimestamp decodes the payload of a timestamp extension in any of its
// three forms.
func UnpackTimestamp(data []byte) (Timestamp, error) {
	var sec int64
	var nsec int64
	switch len(data) {
	case 4:
		sec = int64(binary.BigEndian.Uint32(data))
	case 8:
		v := binary.BigEndian.Uint64(data)
		nsec = int64(v >> 34)
		sec = int64(v & uint64(maxSeconds34))
	case 12:
		nsec = int64(binary.BigEndian.Uint32(data))
		sec = int64(binary.BigEndian.Uint64(data[4:]))
	default:
		return nil, xerrors.Errorf("value: timestamp payload of %d bytes, want 4, 8 or 12", len(data))
	}
	if nsec >= nanosPerSecond {
		return nil, xerrors.Errorf("value: timestamp nanoseconds %d out of range", nsec)
	}
	return timestampValue{sec: sec, nsec: int32(nsec)}, nil
}

// timestampPayload returns the smallest payload that holds the timestamp.
func (t timestampValue) payload() []byte {
	switch {
	case t.nsec == 0 && t.sec >= 0 && t.sec <= maxSeconds32:
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], uint32(t.sec))
		return b[:]
	case t.sec >= 0 && t.sec <= maxSeconds34:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(t.nsec)<<34|uint64(t.sec))
		return b[:]
	default:
		var b [12]byte
		binary.BigEndian.PutUint32(b[:4], uint32(t.nsec))
		binary.BigEndian.PutUint64(b[4:], uint64(t.sec))
		return b[:]
	}
}

func (t timestampValue) Kind() Kind {
	return TIMESTAMP
}

func (t timestampValue) Class() reflect.Type {
	return timestampValueClass
}

func (t timestampValue) Object() interface{} {
	return t.Time()
}

func (t timestampValue) Time() time.Time {
	return time.Unix(t.sec, int64(t.nsec)).UTC()
}

func (t timestampValue) Unix() (int64, int32) {
	return t.sec, t.nsec
}

func (t timestampValue) String() string {
	return t.Time().Format(time.RFC3339Nano)
}

func (t timestampValue) Pack(p Packer) {
	p.PackExt(TimestampExt, t.payload())
}

func (t timestampValue) PrintJSON(out *strings.Builder) {
	out.WriteString(strconv.Quote(t.String()))
}

func (t timestampValue) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(t.String())), nil
}

func (t timestampValue) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	p := MessagePacker(&buf)
	t.Pack(p)
	return buf.Bytes(), p.Error()
}

// Equal reports whether val is a timestamp of the same instant. Equal
// timestamps pack to identical bytes.
func (t timestampValue) Equal(val Value) bool {
	if val == nil || val.Kind() != TIMESTAMP {
		return false
	}
	other, ok := val.(Timestamp)
	if !ok {
		return false
	}
	sec, nsec := other.Unix()
	return t.sec == sec && t.nsec == nsec
}
