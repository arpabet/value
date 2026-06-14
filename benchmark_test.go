/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value_test

import (
	"testing"

	val "go.arpabet.com/value"
)

// BenchmarkPack and BenchmarkUnpack are real Go benchmarks (run with
// `go test -bench=. -benchmem`), replacing the wall-clock timing in
// TestBenchmark. They reuse testCreateMap from api_test.go.

func BenchmarkPack(b *testing.B) {
	m := testCreateMap()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := val.Pack(m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnpack(b *testing.B) {
	data, err := val.Pack(testCreateMap())
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := val.Unpack(data, false); err != nil {
			b.Fatal(err)
		}
	}
}
