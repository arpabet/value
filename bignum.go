/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value

import (
	"encoding/binary"
	"math"
	"math/big"

	"github.com/shopspring/decimal"
	"golang.org/x/xerrors"
)

/*
Portable payloads of the two big-number extensions, the wire format since
v1.5.0:

  - BigInt (ext 1): the two's-complement big-endian bytes of the integer in
    the minimal length that still holds the sign, at least one byte. This is
    java.math.BigInteger.toByteArray() and Avro's convention: 0 is 00, 127 is
    7f, 128 is 0080, -1 is ff, -128 is 80, -129 is ff7f.
  - Decimal (ext 2): a 4-byte big-endian signed scale followed by the
    two's-complement big-endian unscaled integer as above; the value is
    unscaled × 10^-scale. This is java.math.BigDecimal(BigInteger, scale) and
    Avro's decimal layout with the scale carried in the value. 1.10 is
    00000002 6e.

Decoding accepts longer, non-minimal two's-complement encodings too; packing
always writes the minimal form, so Equal values pack identically.

Before v1.5.0 both payloads were Go framings: math/big.Int.GobEncode for
BigInt and shopspring's decimal.MarshalBinary (int32 exponent plus the gob
integer) for Decimal. The two layouts cannot be told apart by looking at the
bytes, so there is no automatic dual read: a program that has to read data
written by an older version sets LegacyBigNumbers while it migrates, or calls
UnpackLegacyBigInt and UnpackLegacyDecimal itself.
*/

// LegacyBigNumbers, when true, makes Unpack, Read and the other decoders read
// the payloads of BigInt (ext 1) and Decimal (ext 2) in the Go layouts that
// versions before v1.5.0 wrote. Packing always writes the portable layouts.
// It exists for migrating stored data and is false by default.
var LegacyBigNumbers = false

// twosComplement returns the minimal two's-complement big-endian encoding of
// x: at least one byte, and the top bit of the first byte is the sign.
func twosComplement(x *big.Int) []byte {
	if x.Sign() >= 0 {
		b := x.Bytes()
		if len(b) == 0 || b[0]&0x80 != 0 {
			return append([]byte{0}, b...)
		}
		return b
	}
	// 2^(8n) + x in n bytes, n large enough to hold |x| and the sign, then
	// drop the sign-extension bytes that carry no information.
	n := (x.BitLen() + 8) / 8
	t := new(big.Int).Lsh(big.NewInt(1), uint(8*n))
	t.Add(t, x)
	b := t.Bytes()
	for len(b) > 1 && b[0] == 0xff && b[1]&0x80 != 0 {
		b = b[1:]
	}
	return b
}

// fromTwosComplement decodes a two's-complement big-endian integer of any
// length of at least one byte.
func fromTwosComplement(b []byte) (*big.Int, error) {
	if len(b) == 0 {
		return nil, xerrors.New("value: empty big integer payload")
	}
	x := new(big.Int).SetBytes(b)
	if b[0]&0x80 != 0 {
		m := new(big.Int).Lsh(big.NewInt(1), uint(8*len(b)))
		x.Sub(x, m)
	}
	return x, nil
}

func bigIntPayload(x *big.Int) []byte {
	return twosComplement(x)
}

func decimalPayload(d decimal.Decimal) []byte {
	// The scale is the negated exponent. -2^31 is the one int32 exponent
	// whose negation is not an int32; it is far beyond MaxDecimalExponent, so
	// such a value never decodes anyway, and it is written with the largest
	// scale.
	scale := -int64(d.Exponent())
	if scale > math.MaxInt32 {
		scale = math.MaxInt32
	}
	unscaled := twosComplement(d.Coefficient())
	b := make([]byte, 4+len(unscaled))
	binary.BigEndian.PutUint32(b, uint32(int32(scale)))
	copy(b[4:], unscaled)
	return b
}

func checkDecimalExponent(exp int64) error {
	if MaxDecimalExponent > 0 && (exp > int64(MaxDecimalExponent) || exp < -int64(MaxDecimalExponent)) {
		return xerrors.Errorf("value: decimal exponent %d exceeds limit %d", exp, MaxDecimalExponent)
	}
	return nil
}

// UnpackBigInt decodes the payload of a BigInt extension in the portable
// two's-complement layout.
func UnpackBigInt(data []byte) (*big.Int, error) {
	return fromTwosComplement(data)
}

// UnpackDecimal decodes the payload of a Decimal extension in the portable
// scale-and-unscaled layout. It rejects a scale whose magnitude exceeds
// MaxDecimalExponent.
func UnpackDecimal(data []byte) (decimal.Decimal, error) {
	if len(data) < 5 {
		return decimal.Decimal{}, xerrors.Errorf("value: decimal payload of %d bytes, want at least 5", len(data))
	}
	scale := int64(int32(binary.BigEndian.Uint32(data)))
	exp := -scale
	if err := checkDecimalExponent(exp); err != nil {
		return decimal.Decimal{}, err
	}
	unscaled, err := fromTwosComplement(data[4:])
	if err != nil {
		return decimal.Decimal{}, err
	}
	return decimal.NewFromBigInt(unscaled, int32(exp)), nil
}

// UnpackLegacyBigInt decodes a BigInt payload written by value before v1.5.0:
// math/big.Int's gob encoding.
func UnpackLegacyBigInt(data []byte) (*big.Int, error) {
	x := new(big.Int)
	if err := x.GobDecode(data); err != nil {
		return nil, xerrors.Errorf("value: legacy big integer payload: %w", err)
	}
	return x, nil
}

// UnpackLegacyDecimal decodes a Decimal payload written by value before
// v1.5.0: shopspring's MarshalBinary, an int32 exponent followed by the gob
// encoding of the coefficient.
func UnpackLegacyDecimal(data []byte) (decimal.Decimal, error) {
	d := decimal.Decimal{}
	if err := d.UnmarshalBinary(data); err != nil {
		return decimal.Decimal{}, xerrors.Errorf("value: legacy decimal payload: %w", err)
	}
	if err := checkDecimalExponent(int64(d.Exponent())); err != nil {
		return decimal.Decimal{}, err
	}
	return d, nil
}
