/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value

import (
	"math"
	"math/big"
	"strconv"

	"github.com/shopspring/decimal"
	"golang.org/x/xerrors"
)

/*
Lossless number conversions (CANONICAL.md, "Numbers into Go fields"). Each
converts a Number into a Go numeric type only when nothing is lost, and fails
otherwise; a codec built with Lossless uses them for every numeric field, and
decimal.Decimal and big.Int fields always do.
*/

// ExactInt returns n as a signed integer of the bit size: n must be integral
// and within range.
func ExactInt(n Number, bits int) (int64, error) {
	i, err := ExactBigInt(n)
	if err != nil {
		return 0, err
	}
	lo := new(big.Int).Lsh(big.NewInt(-1), uint(bits-1))
	hi := new(big.Int).Sub(new(big.Int).Neg(lo), big.NewInt(1))
	if i.Cmp(lo) < 0 || i.Cmp(hi) > 0 {
		return 0, xerrors.Errorf("value: %s overflows int%d", n, bits)
	}
	return i.Int64(), nil
}

// ExactUint returns n as an unsigned integer of the bit size: n must be
// integral, not negative and within range.
func ExactUint(n Number, bits int) (uint64, error) {
	i, err := ExactBigInt(n)
	if err != nil {
		return 0, err
	}
	hi := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	if i.Sign() < 0 || i.Cmp(hi) > 0 {
		return 0, xerrors.Errorf("value: %s overflows uint%d", n, bits)
	}
	return i.Uint64(), nil
}

// ExactBigInt returns n as a big integer: n must be integral.
func ExactBigInt(n Number) (*big.Int, error) {
	switch n.Type() {
	case LONG:
		return big.NewInt(n.Long()), nil
	case BIGINT:
		return new(big.Int).Set(n.BigInt()), nil
	case DOUBLE:
		d := n.Double()
		if math.IsNaN(d) || math.IsInf(d, 0) || d != math.Trunc(d) {
			return nil, xerrors.Errorf("value: double %s is not an integer", DoubleText(d))
		}
		i, _ := new(big.Float).SetFloat64(d).Int(nil)
		return i, nil
	case DECIMAL:
		d := n.Decimal()
		if !d.IsInteger() {
			return nil, xerrors.Errorf("value: decimal %s is not an integer", DecimalText(d))
		}
		return d.BigInt(), nil
	}
	return nil, xerrors.Errorf("value: number type %s", n.Type())
}

// ExactFloat returns n as a float of the bit size, 32 or 64: a double fits a
// float64 always and a float32 when it is one; any other number fits when the
// shortest text of the float reads back as the same number.
func ExactFloat(n Number, bits int) (float64, error) {
	if n.Type() == DOUBLE {
		d := n.Double()
		if bits == 32 && !math.IsNaN(d) && float64(float32(d)) != d {
			return 0, xerrors.Errorf("value: double %s is not a float32", DoubleText(d))
		}
		return d, nil
	}
	exact, err := ExactDecimal(n)
	if err != nil {
		return 0, err
	}
	f, err := strconv.ParseFloat(exact.String(), bits)
	if err != nil {
		return 0, xerrors.Errorf("value: %s is not a float%d: %w", n, bits, err)
	}
	back, err := decimal.NewFromString(strconv.FormatFloat(f, 'g', -1, bits))
	if err != nil || !back.Equal(exact) {
		return 0, xerrors.Errorf("value: %s is not exactly a float%d", n, bits)
	}
	return f, nil
}

// ExactDecimal returns n as a decimal: a double through its shortest text,
// which must be finite.
func ExactDecimal(n Number) (decimal.Decimal, error) {
	switch n.Type() {
	case LONG:
		return decimal.NewFromInt(n.Long()), nil
	case BIGINT:
		return decimal.NewFromBigInt(n.BigInt(), 0), nil
	case DOUBLE:
		d := n.Double()
		if math.IsNaN(d) || math.IsInf(d, 0) {
			return decimal.Decimal{}, xerrors.Errorf("value: double %s is not a decimal", DoubleText(d))
		}
		return decimal.NewFromString(strconv.FormatFloat(d, 'g', -1, 64))
	case DECIMAL:
		return n.Decimal(), nil
	}
	return decimal.Decimal{}, xerrors.Errorf("value: number type %s", n.Type())
}
