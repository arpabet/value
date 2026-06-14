/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package value_test

import (
	"bytes"
	"testing"

	val "go.arpabet.com/value"
)

// seedCorpus returns valid encodings plus a few hostile headers, used to seed
// both fuzz targets.
func seedCorpus() [][]byte {
	values := []val.Value{
		val.Null,
		val.Boolean(true),
		val.Long(123),
		val.Double(-12.34),
		val.Utf8("hello"),
		val.Raw([]byte{0, 1, 2}, false),
		val.ImmutableList([]val.Value{val.Long(1), val.Utf8("x")}),
		val.ImmutableMapOf(map[string]val.Value{"a": val.Long(1), "b": val.Boolean(false)}),
	}
	var seeds [][]byte
	for _, v := range values {
		if mp, err := val.Pack(v); err == nil {
			seeds = append(seeds, mp)
		}
	}
	seeds = append(seeds,
		[]byte{},
		[]byte{0xc1},                         // never-used code
		[]byte{0xdd, 0xff, 0xff, 0xff, 0xff}, // array32 claiming ~4.29e9 elements
		[]byte{0xdf, 0xff, 0xff, 0xff, 0xff}, // map32 claiming ~4.29e9 entries
		[]byte{0xdb, 0xff, 0xff, 0xff, 0xff}, // str32 claiming ~4.29e9 bytes
	)
	return seeds
}

// FuzzUnpack feeds arbitrary bytes to Unpack. It must never panic, hang, or
// exhaust memory; any value that decodes must re-pack and re-decode cleanly.
func FuzzUnpack(f *testing.F) {
	for _, s := range seedCorpus() {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		v, err := val.Unpack(data, false)
		if err != nil {
			return
		}
		mp, err := val.Pack(v)
		if err != nil {
			t.Fatalf("pack of decoded value failed: %v", err)
		}
		if _, err := val.Unpack(mp, false); err != nil {
			t.Fatalf("re-unpack of packed value failed: %v", err)
		}
	})
}

// FuzzRead exercises the streaming reader path, which must also never panic on
// arbitrary input.
func FuzzRead(f *testing.F) {
	for _, s := range seedCorpus() {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = val.Read(bytes.NewReader(data))
	})
}
