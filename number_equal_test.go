/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package value_test

import (
	"math/big"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	val "go.arpabet.com/value"
)

// numberSamples returns at least one value of every Number type plus a few
// close neighbors, used to exercise the equality contract.
func numberSamples() []val.Number {
	return []val.Number{
		val.Long(0),
		val.Long(1),
		val.Long(-123),
		val.Double(0),
		val.Double(1),         // 1.0, numerically equal to Long(1) but a different type
		val.Double(1.0000005), // within PrecisionLevel of the next sample, but not bit-equal
		val.Double(1.0000006),
		val.BigInt(big.NewInt(1)),
		val.BigInt(big.NewInt(123)),
		val.Decimal(decimal.New(1, 0)),   // 1
		val.Decimal(decimal.New(123, 0)), // 123
	}
}

// TestNumberEqualReflexive verifies every value equals itself, including NaN.
// Reflexivity is required for round-tripping (Pack then Unpack then Equal) and
// for use as map/set members.
func TestNumberEqualReflexive(t *testing.T) {
	for _, n := range append(numberSamples(), val.Nan) {
		require.True(t, n.Equal(n), "value must equal itself: %s", n.String())
	}
}

// TestNumberEqualSymmetric verifies a.Equal(b) == b.Equal(a) for all pairs.
// This is the regression guard for the old truncating comparison, where
// Long(1).Equal(Double(1.5)) was true but the reverse was false.
func TestNumberEqualSymmetric(t *testing.T) {
	samples := append(numberSamples(), val.Nan)
	for i, a := range samples {
		for j, b := range samples {
			require.Equal(t, a.Equal(b), b.Equal(a),
				"symmetry violated for [%d]%s vs [%d]%s", i, a.String(), j, b.String())
		}
	}
}

// TestNumberEqualImpliesEqualBytes is the core content-addressing invariant:
// if two numbers are Equal they must serialize to identical bytes.
func TestNumberEqualImpliesEqualBytes(t *testing.T) {
	samples := append(numberSamples(), val.Nan)
	for i, a := range samples {
		for j, b := range samples {
			if !a.Equal(b) {
				continue
			}
			pa, err := val.Pack(a)
			require.NoError(t, err)
			pb, err := val.Pack(b)
			require.NoError(t, err)
			require.Equal(t, pa, pb,
				"Equal but different bytes: [%d]%s vs [%d]%s", i, a.String(), j, b.String())
		}
	}
}

// TestNumberEqualCrossTypeDistinct verifies that the value 1 represented as a
// long, double, big.Int and decimal are all mutually distinct, because each
// packs to different bytes.
func TestNumberEqualCrossTypeDistinct(t *testing.T) {
	one := []val.Number{
		val.Long(1),
		val.Double(1),
		val.BigInt(big.NewInt(1)),
		val.Decimal(decimal.New(1, 0)),
	}
	for i, a := range one {
		for j, b := range one {
			if i == j {
				require.True(t, a.Equal(b), "same instance must be Equal: %s", a.String())
			} else {
				require.False(t, a.Equal(b),
					"different number types must not be Equal: %s vs %s", a.String(), b.String())
			}
		}
	}
}

// TestNumberEqualNoTruncation is the direct regression for the asymmetric,
// value-truncating long comparison.
func TestNumberEqualNoTruncation(t *testing.T) {
	require.False(t, val.Long(1).Equal(val.Double(1.5)))
	require.False(t, val.Double(1.5).Equal(val.Long(1)))
}

// TestDoubleEqualExact is the regression for the old tolerance-based double
// comparison: two doubles 1e-7 apart used to be Equal yet hashed differently.
func TestDoubleEqualExact(t *testing.T) {
	require.False(t, val.Double(1.0000005).Equal(val.Double(1.0000006)))
	require.True(t, val.Double(1.5).Equal(val.Double(1.5)))
}

// TestApproxEqual verifies the opt-in fuzzy, cross-type comparison preserves the
// old behavior for callers who explicitly want it.
func TestApproxEqual(t *testing.T) {
	require.True(t, val.ApproxEqual(val.Long(1), val.Double(1.0)))
	require.True(t, val.ApproxEqual(val.Double(1.0000005), val.Double(1.0000006)))
	require.False(t, val.ApproxEqual(val.Long(1), val.Double(1.5)))
	require.False(t, val.ApproxEqual(val.Nan, val.Nan))
}
