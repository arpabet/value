/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package value_test

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	val "go.arpabet.com/value"
)

// --- #1: array header for lists > 65535 must use array32 (0xdd), not array16 ---

func TestPackLargeListUsesArray32(t *testing.T) {
	const n = 70000 // > math.MaxUint16 forces the 32-bit array header
	vals := make([]val.Value, n)
	for i := range vals {
		vals[i] = val.Long(int64(i & 0x3f))
	}
	list := val.ImmutableList(vals)

	mp, err := val.Pack(list)
	require.NoError(t, err)
	require.Equal(t, byte(0xdd), mp[0], "lists > 65535 must use the array32 header (0xdd)")

	out, err := val.Unpack(mp, false)
	require.NoError(t, err)
	require.Equal(t, val.LIST, out.Kind())
	require.Equal(t, n, out.(val.List).Len())
	require.True(t, list.Equal(out))
}

func TestPackListBoundaryUsesArray16(t *testing.T) {
	const n = 65535 // == math.MaxUint16 still fits the 16-bit array header
	vals := make([]val.Value, n)
	for i := range vals {
		vals[i] = val.Long(1)
	}
	mp, err := val.Pack(val.ImmutableList(vals))
	require.NoError(t, err)
	require.Equal(t, byte(0xdc), mp[0], "lists == 65535 use the array16 header (0xdc)")
}

// --- #2: the packer must surface io.Writer errors instead of swallowing them ---

type errAfterWriter struct {
	okWrites int
	seen     int
}

func (w *errAfterWriter) Write(p []byte) (int, error) {
	w.seen++
	if w.seen > w.okWrites {
		return 0, errors.New("write failed")
	}
	return len(p), nil
}

func TestPackerPropagatesWriteError(t *testing.T) {
	list := val.ImmutableList([]val.Value{val.Long(1), val.Long(2), val.Long(3)})
	// header write succeeds, the first element write fails.
	err := val.Write(&errAfterWriter{okWrites: 1}, list)
	require.Error(t, err, "Write must surface io.Writer failures")
}

// --- #3: a map with integer keys followed by a string key must decode without
// losing/corrupting entries, and the resulting string-keyed map must be sorted
// (stringified-int keys like "10" do not sort lexicographically with "2"). ---

func TestUnpackPromotedIntStringMap(t *testing.T) {
	// {0:0, 1:1, ... 10:10, "k":99} as MessagePack (fixmap, 12 entries).
	blob := []byte{
		0x8c, // fixmap, 12 entries
		0x00, 0x00,
		0x01, 0x01,
		0x02, 0x02,
		0x03, 0x03,
		0x04, 0x04,
		0x05, 0x05,
		0x06, 0x06,
		0x07, 0x07,
		0x08, 0x08,
		0x09, 0x09,
		0x0a, 0x0a, // key 10, value 10 — "10" breaks lexicographic order with "2"
		0xa1, 'k', 0x63, // "k": 99
	}

	out, err := val.Unpack(blob, false)
	require.NoError(t, err)
	require.Equal(t, val.MAP, out.Kind())

	m := out.(val.Map)
	require.Equal(t, 12, m.Len())
	require.Equal(t, int64(0), m.GetNumber("0").Long())
	require.Equal(t, int64(2), m.GetNumber("2").Long())
	require.Equal(t, int64(10), m.GetNumber("10").Long()) // fails if the map is left unsorted
	require.Equal(t, int64(99), m.GetNumber("k").Long())
}

// --- #4.4 and #4.5: reading from a plain io.Reader (not an io.ByteReader) and
// from a reader that returns short reads must work without panicking. ---

// plainReader is an io.Reader that does NOT implement io.ByteReader.
type plainReader struct {
	data []byte
	pos  int
}

func (r *plainReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// oneByteReader yields at most one byte per Read, forcing partial reads.
type oneByteReader struct {
	data []byte
	pos  int
}

func (r *oneByteReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	if len(p) == 0 {
		return 0, nil
	}
	p[0] = r.data[r.pos]
	r.pos++
	return 1, nil
}

func TestReadFromNonByteReader(t *testing.T) {
	orig := val.Utf8("hello world")
	mp, err := val.Pack(orig)
	require.NoError(t, err)

	out, err := val.Read(&plainReader{data: mp})
	require.NoError(t, err)
	require.True(t, orig.Equal(out))
}

func TestReadHandlesShortReads(t *testing.T) {
	orig := val.Long(2147483647) // 5-byte encoding exercises multi-byte header reads
	mp, err := val.Pack(orig)
	require.NoError(t, err)

	out, err := val.Read(&oneByteReader{data: mp})
	require.NoError(t, err)
	require.True(t, orig.Equal(out))
}

// --- Unseal must reject inputs shorter than the 24-byte nonce instead of panicking ---

func TestUnsealRejectsShortInput(t *testing.T) {
	var pub, priv [32]byte
	_, err := val.Unseal([]byte{1, 2, 3}, &pub, &priv)
	require.Error(t, err)
}
