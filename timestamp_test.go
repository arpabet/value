/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value_test

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	val "go.arpabet.com/value"
)

func TestTimestampForms(t *testing.T) {
	cases := []struct {
		name string
		v    val.Timestamp
		hex  string
	}{
		{"epoch", val.UnixTime(0, 0), "d6ff00000000"},
		{"seconds", val.UnixTime(1793345319, 0), "d6ff6ae44727"},
		{"millis", val.UnixTime(1793345319, 75000000), "d7ff11e1a3006ae44727"},
		{"nanos", val.UnixTime(1793345319, 123456789), "d7ff1d6f34546ae44727"},
		{"largest-32", val.UnixTime(1<<32-1, 0), "d6ffffffffff"},
		{"first-64", val.UnixTime(1<<32, 0), "d7ff0000000100000000"},
		{"largest-64", val.UnixTime(1<<34-1, 999999999), "d7ffee6b27ffffffffff"},
		{"first-96", val.UnixTime(1<<34, 0), "c70cff000000000000000400000000"},
		{"before-epoch", val.UnixTime(-1, 0), "c70cff00000000ffffffffffffffff"},
		{"before-epoch-half", val.UnixTime(-1, 500000000), "c70cff1dcd6500ffffffffffffffff"},
		{"zero-time", val.Time(time.Time{}), "c70cff00000000fffffff1886e0900"},
	}
	for _, c := range cases {
		require.Equal(t, c.hex, val.Hex(c.v), c.name)
		b, err := hex.DecodeString(c.hex)
		require.NoError(t, err)
		back, err := val.Unpack(b, false)
		require.NoError(t, err, c.name)
		require.Equal(t, val.TIMESTAMP, back.Kind(), c.name)
		require.True(t, c.v.Equal(back), "%s: %s != %s", c.name, c.v, back)
		require.Equal(t, c.hex, val.Hex(back), c.name)
	}
}

func TestTimestampValue(t *testing.T) {
	cet := time.FixedZone("CET", 7200)
	instant := time.Date(2026, 10, 30, 7, 28, 39, 75000000, time.UTC)
	ts := val.Time(instant.In(cet))

	require.Equal(t, val.TIMESTAMP, ts.Kind())
	require.Equal(t, "TIMESTAMP", ts.Kind().String())
	require.Equal(t, "value.timestampValue", ts.Class().String())
	require.Equal(t, "2026-10-30T07:28:39.075Z", ts.String())
	require.Equal(t, "\"2026-10-30T07:28:39.075Z\"", val.Jsonify(ts))
	js, err := ts.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, "\"2026-10-30T07:28:39.075Z\"", string(js))
	require.True(t, instant.Equal(ts.Time()))
	require.Equal(t, time.UTC, ts.Time().Location())
	require.Equal(t, instant, ts.Object())
	sec, nsec := ts.Unix()
	require.Equal(t, int64(1793345319), sec)
	require.Equal(t, int32(75000000), nsec)

	// The location is not part of the value.
	require.True(t, ts.Equal(val.Time(instant)))
	require.Equal(t, val.Hex(val.Time(instant)), val.Hex(ts))
	// Other kinds and other instants are not equal.
	require.False(t, ts.Equal(val.Long(1793345319)))
	require.False(t, ts.Equal(val.UnixTime(1793345319, 75000001)))
	require.False(t, ts.Equal(nil))

	bin, err := ts.MarshalBinary()
	require.NoError(t, err)
	require.Equal(t, "d7ff11e1a3006ae44727", hex.EncodeToString(bin))

	// Whole seconds render without a fraction.
	require.Equal(t, "2026-10-30T07:28:39Z", val.UnixTime(1793345319, 0).String())
	// Nanoseconds are normalized into the second.
	require.True(t, val.UnixTime(10, 2500000000).Equal(val.UnixTime(12, 500000000)))
	require.True(t, val.UnixTime(10, -1).Equal(val.UnixTime(9, 999999999)))
}

func TestParseTimestamp(t *testing.T) {
	ts, err := val.ParseTimestamp("2026-10-30T09:28:39.075+02:00")
	require.NoError(t, err)
	require.Equal(t, "2026-10-30T07:28:39.075Z", ts.String())
	ts, err = val.ParseTimestamp("1969-12-31T23:59:59.5Z")
	require.NoError(t, err)
	require.Equal(t, "c70cff1dcd6500ffffffffffffffff", val.Hex(ts))
	_, err = val.ParseTimestamp("yesterday")
	require.Error(t, err)
}

func TestUnpackTimestampRejects(t *testing.T) {
	bad := []struct {
		name string
		hex  string
	}{
		{"wrong-length-2", "d5ff0000"},
		{"wrong-length-16", "d8ff00000000000000000000000000000000"},
		{"wrong-length-5", "c705ff0000000000"},
		{"nanos-64", "d7ffee6b2800" + "00000000"},      // 1e9 << 34
		{"nanos-96", "c70cff3b9aca000000000000000000"}, // nanos 1e9
	}
	for _, c := range bad {
		b, err := hex.DecodeString(c.hex)
		require.NoError(t, err, c.name)
		_, err = val.Unpack(b, false)
		require.Error(t, err, c.name)
	}
	// Non-minimal forms of a representable instant decode, and re-pack
	// canonically.
	b, err := hex.DecodeString("c70cff00000000000000006ae44727") // 96-bit form of a 32-bit instant
	require.NoError(t, err)
	v, err := val.Unpack(b, false)
	require.NoError(t, err)
	require.Equal(t, "d6ff6ae44727", val.Hex(v))
}

func TestTimestampInCollections(t *testing.T) {
	ts := val.UnixTime(1793345319, 0)
	m := val.ImmutableMapOf(map[string]val.Value{"when": ts, "n": val.Long(1)})
	mp, err := val.Pack(m)
	require.NoError(t, err)
	back, err := val.Unpack(mp, false)
	require.NoError(t, err)
	require.True(t, m.Equal(back))
	got, ok := val.As[val.Timestamp](back.(val.Map).Get("when"))
	require.True(t, ok)
	require.True(t, ts.Equal(got))
	require.Equal(t, `{"n": 1,"when": "2026-10-30T07:28:39Z"}`, val.Jsonify(back))

	l := val.ImmutableList([]val.Value{ts, val.Null})
	mp, err = val.Pack(l)
	require.NoError(t, err)
	back, err = val.Unpack(mp, false)
	require.NoError(t, err)
	require.True(t, l.Equal(back))
}

func TestTimestampStruct(t *testing.T) {
	type event struct {
		When val.Timestamp `tag:"1"`
		Name val.String    `tag:"2"`
	}
	src := &event{When: val.UnixTime(1793345319, 75000000), Name: val.Utf8("x")}
	b, err := val.PackStruct(src)
	require.NoError(t, err)
	var dst event
	require.NoError(t, val.UnpackStruct(b, &dst, false))
	require.True(t, src.When.Equal(dst.When))
	require.True(t, src.Name.Equal(dst.Name))
}

func TestTimestampCodecProfiles(t *testing.T) {
	type record struct {
		At   time.Time  `value:"at" json:"at" msgpack:"at"`
		Seen *time.Time `value:"seen,omitempty" json:"seen,omitempty" msgpack:"seen,omitempty"`
	}
	at := time.Date(2026, 10, 30, 7, 28, 39, 123456789, time.UTC)
	for _, codec := range []val.ValueCodec{val.DefaultValueCodec(), val.JSONValueCodec(), val.MsgpackValueCodec()} {
		v, err := codec.Marshal(record{At: at})
		require.NoError(t, err)
		require.Equal(t, val.TIMESTAMP, v.(val.Map).Get("at").Kind())
		require.Equal(t, 1, v.(val.Map).Len())
		var dst record
		require.NoError(t, codec.Unmarshal(v, &dst))
		require.True(t, at.Equal(dst.At))
		require.Equal(t, 123456789, dst.At.Nanosecond())
		require.Nil(t, dst.Seen)
	}
}
