# Canonical form

This document specifies the canonical serialization produced by `value.Pack` (and
`value.Write`/`value.Hash`). It is the contract that makes the library suitable
for **content addressing, hashing, and cryptographic signing**: equal values
produce identical bytes, so they produce identical hashes and verifiable
signatures. That property is what lets `value` back Merkle structures, content
identifiers, and reproducible signatures in **blockchain** and other distributed
protocols.

The canonical form is the **MessagePack byte stream**. `String()` and JSON output
(`Jsonify`, `MarshalJSON`) are human-facing and are **not** canonical — do not hash
them.

This document is normative for every implementation of the format, in any
language. The vectors in [`testdata/canonical.json`](testdata/canonical.json)
are the conformance suite: an implementation must produce exactly those bytes
for those values and decode them back to equal values. `vectors_test.go` runs
them here; `canonical_test.go`, `number_test.go`, `bignum_test.go` and
`timestamp_test.go` pin further encodings.

## Equality contract

For any two values `a` and `b` of the same `Kind`:

> `a.Equal(b)` ⟺ `Pack(a)` equals `Pack(b)`

Consequences:

- **Numbers do not coerce across types.** `Long(1)`, `Double(1.0)`, `BigInt(1)` and
  `Decimal(1)` are four distinct values with four distinct encodings; none are
  `Equal`. For an approximate, cross-type numeric comparison use `ApproxEqual`,
  which is explicitly **not** canonical.
- **Doubles compare by exact IEEE-754 bits**, so equality is reflexive even for
  `NaN` (a value always equals itself, which round-tripping relies on).
- **Decimals compare by exact representation** (coefficient + exponent), so `1.0`
  and `1.00` are distinct, matching their distinct bytes.
- **Timestamps compare by instant** (seconds and nanoseconds). The location a
  `time.Time` carried is not part of the value, so the same instant in two zones
  is one timestamp with one encoding.

One subtlety this invariant does **not** cover: a dense `List` and a `SparseList`
holding the same elements are `Equal` at the value level but encode differently
(array vs. integer-keyed map, see below). Choose the collection type deliberately
when the bytes are hashed.

## Type → encoding

| Kind | Go constructor | MessagePack | Example (hex) |
|---|---|---|---|
| NULL | `Null` | `nil` (`0xc0`) | `c0` |
| BOOL | `Boolean(true/false)` | `true`/`false` | `c3` / `c2` |
| NUMBER / long | `Long(n)` | minimal int family | `Long(0)`→`00`, `Long(128)`→`cc80` |
| NUMBER / double | `Double(f)` | always `float64` (`0xcb`) | `Double(1.5)`→`cb3ff8000000000000` |
| NUMBER / bigint | `BigInt(x)` | ext type `1`, two's complement | `BigInt(123)`→`d4017b` |
| NUMBER / decimal | `Decimal(d)` | ext type `2`, scale + unscaled | `Decimal(1.10)`→`c70502000000026e` |
| STRING / utf8 | `Utf8(s)` | `str` family | `Utf8("hello")`→`a568656c6c6f` |
| STRING / raw | `Raw(b)` | `bin` family | `Raw(0x010203)`→`c403010203` |
| TIMESTAMP | `Time(t)`, `UnixTime(sec, nsec)` | ext type `-1`, smallest form | `2026-10-30T07:28:39Z`→`d6ff6ae44727` |
| LIST / dense | `MutableList`, `ImmutableList` | `array` | `[1,2,3]`→`93010203` |
| LIST / sparse | `SparseList` | integer-keyed `map` | `{0:10,2:20}`→`82000a0214` |
| MAP | `SortedMap`, `ImmutableMap` | string-keyed `map`, sorted | `{a:1,b:2}`→`82a16101a16202` |
| UNKNOWN | `Unknown(tagAndData)` | any other ext, passed through | `d40701` |

### Integers (LONG)

Encoded in the **smallest** MessagePack integer form that fits the `int64` value
(positive/negative fixint, then `int8/16/32/64` or `uint8/16/32/64`). Boundary
encodings are pinned in `number_test.go`.

### Floats (DOUBLE)

Always encoded as 64-bit `float64` (`0xcb` + 8 bytes, big-endian IEEE-754) — never
`float32`. `NaN` uses Go's canonical quiet-NaN bit pattern
(`cb7ff8000000000001`).

### Big integers and decimals (extensions)

Both payloads are language-neutral since v1.5.0 (they were Go framings before;
see *Reading data written before v1.5.0* below).

- `BigInt` → ext type **1** (`BigIntExt`). The payload is the **two's-complement
  big-endian** encoding of the integer in the **minimal length that still holds
  the sign**, and at least one byte. This is `java.math.BigInteger.toByteArray()`
  and Avro's convention.

  | value | payload | bytes |
  |---|---|---|
  | 0 | `00` | `d40100` |
  | 127 | `7f` | `d4017f` |
  | 128 | `0080` | `d5010080` |
  | -1 | `ff` | `d401ff` |
  | -128 | `80` | `d40180` |
  | -129 | `ff7f` | `d501ff7f` |
  | 2^64 | `010000000000000000` | `c70901010000000000000000` |

- `Decimal` → ext type **2** (`DecimalExt`). The payload is a **4-byte big-endian
  signed scale** followed by the **two's-complement big-endian unscaled integer**
  as above; the value is `unscaled × 10^−scale`. This is
  `java.math.BigDecimal(BigInteger unscaled, int scale)` and Avro's decimal
  layout with the scale carried in the value. The scale is the negated shopspring
  exponent: `1.10` is coefficient 110 with exponent −2, so scale 2.

  | value | payload | bytes |
  |---|---|---|
  | 0 | `00000000 00` | `c705020000000000` |
  | 1.10 | `00000002 6e` | `c70502000000026e` |
  | 123.45 | `00000002 3039` | `c70602000000023039` |
  | -1.045 | `00000003 fbeb` | `c7060200000003fbeb` |
  | 1e3 | `fffffffd 01` | `c70502fffffffd01` |

  Trailing zeros are significant: `1`, `1.0` and `1.00` have scales 0, 1 and 2
  and distinct bytes, which is what makes them distinct values. The one int32
  exponent whose negation is not an int32, −2^31, is written with scale 2^31−1;
  it lies far beyond `MaxDecimalExponent`, so no such value decodes anyway.

Decoding accepts longer, sign-extended two's-complement encodings and
re-encodes them minimally, so `Equal` values always pack identically. It rejects
an empty integer, a decimal payload shorter than five bytes, and a decimal whose
absolute scale exceeds `MaxDecimalExponent` (see Limits), to prevent later
operations from materializing an enormous integer.

#### Reading data written before v1.5.0

Before v1.5.0 the BigInt payload was `math/big.Int.GobEncode` (a version/sign
byte followed by the magnitude) and the Decimal payload was shopspring's
`decimal.MarshalBinary` (a 4-byte exponent followed by that gob integer). The old
and new layouts cannot be told apart from the bytes (gob's version byte `02` is
also a valid two's-complement first byte), so the decoder never guesses. A program
that holds stored data in the old layout sets `value.LegacyBigNumbers = true`
while it migrates (decoders then read the old payloads; `Pack` always writes the
new ones, so a migration is `Unpack` then `Pack`), or calls `UnpackLegacyBigInt`
and `UnpackLegacyDecimal` itself.

### Timestamps (extension -1)

A `Timestamp` is an instant: `sec` seconds since the Unix epoch and `nsec`
nanoseconds within that second, `0 ≤ nsec < 10^9`. It encodes as MessagePack's
predefined timestamp extension, type **−1** (`TimestampExt`, the byte `ff`), in
the **smallest of its three forms** that holds the instant:

| form | when | layout | example |
|---|---|---|---|
| 32-bit | `nsec = 0` and `0 ≤ sec < 2^32` | fixext 4: `d6 ff` + uint32 seconds | `1970-01-01T00:00:00Z`→`d6ff00000000` |
| 64-bit | `0 ≤ sec < 2^34` | fixext 8: `d7 ff` + uint64 of `nsec << 34 \| sec` | `2026-10-30T07:28:39.075Z`→`d7ff11e1a3006ae44727` |
| 96-bit | otherwise | ext 8: `c7 0c ff` + uint32 nanoseconds + int64 seconds | `1969-12-31T23:59:59.5Z`→`c70cff1dcd6500ffffffffffffffff` |

Decoding accepts any of the three forms for any instant and rejects other payload
lengths and `nsec ≥ 10^9`. `time.Time` fields marshal to this kind and read back
in UTC; `Unmarshal` also still accepts an integer of Unix milliseconds, the
encoding of `time.Time` before v1.5.0. In JSON and `String()` a timestamp renders
as RFC 3339 in UTC with the fraction trimmed (`2026-10-30T07:28:39.075Z`).

### Unknown extensions

An extension of any other type decodes as `UNKNOWN` (`Unknown(tagAndData)`) and
re-encodes byte for byte, so a reader never drops an extension it does not know.

### Strings vs. binary

`Utf8` encodes as the `str` family; `Raw` encodes as the `bin` family. They are
distinct kinds of `STRING` and never interchangeable in the canonical form.

### Lists

- A **dense** list (`MutableList`/`ImmutableList`) encodes as a MessagePack
  `array`.
- A **sparse** list (`SparseList`) encodes as a MessagePack `map` whose keys are
  the integer indices. On decode, a map whose keys are all non-negative `LONG`
  values within `MaxParseCollectionLen` is reconstructed as a `SparseList`;
  otherwise it is a string-keyed map.

### Maps

String-keyed. Entries are **sorted by key** using Go's byte-wise string ordering,
ascending, before encoding. Encoding therefore depends only on the contents, not
on insertion order. Note this is **lexicographic**, so stringified integer keys do
not sort numerically (`"10"` precedes `"2"`).

## JSON

`FormatJSON(v, mode)` / `AppendJSON` write a value as JSON text and `ParseJSON`
reads it back. Plain JSON has numbers, strings, booleans, null, arrays and
objects; every other kind is written as an object with a single **wrapper key**:

| Kind | Wrapper | Example |
|---|---|---|
| LONG | `$long`, decimal digits | `{"$long":"5"}` |
| DOUBLE | `$double`, shortest digits as JavaScript prints them; `NaN`, `Infinity`, `-Infinity`; `-0` keeps its sign | `{"$double":"1.5"}` |
| BIGINT | `$bigint`, decimal digits | `{"$bigint":"-7"}` |
| DECIMAL | `$decimal`, the exact digits: `1.10` keeps its trailing zero, an exponent above 0 or more than six leading zeros is written as `<coefficient>e<exponent>` | `{"$decimal":"1.10"}`, `{"$decimal":"1e400"}` |
| binary | `$bin`, standard base64 with padding | `{"$bin":"AQI="}` |
| TIMESTAMP | `$timestamp`, RFC 3339 in UTC with the fraction trimmed | `{"$timestamp":"2026-10-30T07:28:39.075Z"}` |
| sparse list | `$sparse`, an object of index → item in index order | `{"$sparse":{"2":null,"10":true}}` |
| UNKNOWN | `$ext`, the payload and the signed extension type | `{"$ext":{"base64":"AQ==","type":-7}}` |
| a map whose only key is a wrapper key | `$map`, the map itself, read literally | `{"$map":{"$bin":"x"}}` |

Two modes:

- **Canonical** (`JSONCanonical`) wraps every kind above, so `ParseJSON` reads
  back the same kinds and digits: the canonical MessagePack bytes of the value
  are unchanged by the round trip, and the text is a function of the value (map
  keys sorted bytewise, no insignificant space).
- **Relaxed** (`JSONRelaxed`) is for people and plain JSON tools: LONG, finite
  DOUBLE and DECIMAL are bare JSON numbers with their exact digits; the other
  kinds are wrapped as in canonical mode.

`ParseJSON` reads either mode. A **bare JSON number carries no type and reads as
a DECIMAL** built from its exact text, integers included (`1` is Decimal 1,
`1.10` keeps both digits), so a field never changes kind with the size of its
number. An object is a wrapper only when the wrapper key is its only key; a
malformed wrapper fails instead of reading as a map. Duplicate keys, invalid
UTF-8, lone surrogates, raw control characters, text after the value and numbers
longer than `MaxDecimalExponent` characters fail; the decoding limits below
apply. Strings are written with only the escapes JSON requires (`"`, `\`, and
control characters); a string that is not valid UTF-8 has no JSON form.

`Jsonify` / `PrintJSON` keep their earlier, human-oriented rendering (below),
which value-rpc's JSON wire codec speaks.

## Numbers into Go fields

A codec built with `Lossless()` sets a numeric Go field only when the number
survives the conversion, and fails naming the field otherwise (the default
codec truncates and wraps, as it always has):

| Into | Allowed when |
|---|---|
| `int8` … `int64`, `int` | the number is integral and within range |
| `uint8` … `uint64`, `uint` | the number is integral, not negative and within range |
| `float64` | a DOUBLE always; another number when the shortest text of the float reads back as the same number (`1.10` → 1.1, but not `0.12345678901234567890` or 2^53 + 1) |
| `float32` | the same at 32 bits; a DOUBLE only when it is a float32 |
| `decimal.Decimal` | always; a DOUBLE through its shortest text, and it must be finite (every codec) |
| `big.Int`, `*big.Int` | the number is integral (every codec) |

`decimal.Decimal` and `big.Int` fields marshal as DECIMAL and BIGINT in every
codec. `ExactInt`, `ExactUint`, `ExactFloat`, `ExactDecimal` and `ExactBigInt`
are the conversions themselves. `AllowMissing()` makes a codec leave a field
whose key is absent at its zero value, which a reader of a struct that gained
fields needs for values written before.

## Decoding limits

To stay safe on hostile input, decoding enforces configurable bounds (package
vars in `limits.go`; set any to `0` to disable):

| Var | Default | Guards against |
|---|---|---|
| `MaxParseDepth` | 1000 | stack exhaustion from deep nesting |
| `MaxParseCollectionLen` | 16,777,216 | oversized list/map pre-allocation |
| `MaxParseByteLen` | 1,073,741,824 | oversized string/binary/ext allocation |
| `MaxDecimalExponent` | 1,048,576 | decimals (by absolute scale) that explode on later arithmetic |

## What is *not* canonical

- **`Jsonify` / `PrintJSON` / `String()`** output is for humans (lossless JSON
  is `FormatJSON`, above). Its formatting is affected by
  package vars (`Base64Prefix`, `DecimalExpDelim`, `UnknownPrefix`); changing them
  changes that text but never the MessagePack bytes from `Pack`.
- **Mutability.** Only the `Immutable*` types are safe to share freely. The mutable
  types (`MutableList`/`SortedMap`/`SparseList`) may share backing arrays when
  `AllowFastAppends` is true (the default); treat a value derived from one as
  owning its parent's storage and do not mutate across goroutines. This does not
  affect the bytes of any single value, only aliasing.
