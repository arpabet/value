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

Golden vectors for everything below live in `canonical_test.go`; per-number
encodings are additionally locked in `number_test.go`.

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
| NUMBER / bigint | `BigInt(x)` | ext tag `1` | `BigInt(123)`→`d501027b` |
| NUMBER / decimal | `Decimal(d)` | ext tag `2` | — |
| STRING / utf8 | `Utf8(s)` | `str` family | `Utf8("hello")`→`a568656c6c6f` |
| STRING / raw | `Raw(b)` | `bin` family | `Raw(0x010203)`→`c403010203` |
| LIST / dense | `MutableList`, `ImmutableList` | `array` | `[1,2,3]`→`93010203` |
| LIST / sparse | `SparseList` | integer-keyed `map` | `{0:10,2:20}`→`82000a0214` |
| MAP | `SortedMap`, `ImmutableMap` | string-keyed `map`, sorted | `{a:1,b:2}`→`82a16101a16202` |

### Integers (LONG)

Encoded in the **smallest** MessagePack integer form that fits the `int64` value
(positive/negative fixint, then `int8/16/32/64` or `uint8/16/32/64`). Boundary
encodings are pinned in `number_test.go`.

### Floats (DOUBLE)

Always encoded as 64-bit `float64` (`0xcb` + 8 bytes, big-endian IEEE-754) — never
`float32`. `NaN` uses Go's canonical quiet-NaN bit pattern
(`cb7ff8000000000001`).

### Big integers and decimals (extensions)

- `BigInt` → ext tag **1** (`BigIntExt`), payload = `math/big.Int.GobEncode`
  (one sign/version byte followed by the big-endian magnitude).
- `Decimal` → ext tag **2** (`DecimalExt`), payload = shopspring
  `decimal.MarshalBinary` (4-byte big-endian `int32` exponent followed by the
  gob-encoded coefficient).

> ⚠️ **Cross-language caveat.** Both payloads are Go-/library-specific framings,
> not language-neutral. A non-Go reader must reimplement them, and the bytes are
> coupled to those libraries' encodings. Replacing them with a portable
> sign+magnitude / (sign, coefficient, exponent) framing is a candidate for a
> future **major** version (it changes the wire format and is therefore breaking);
> it is intentionally **not** changed here.

Decoding rejects a decimal whose absolute exponent exceeds `MaxDecimalExponent`
(see Limits) to prevent later operations from materializing an enormous integer.

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

## Decoding limits

To stay safe on hostile input, decoding enforces configurable bounds (package
vars in `limits.go`; set any to `0` to disable):

| Var | Default | Guards against |
|---|---|---|
| `MaxParseDepth` | 1000 | stack exhaustion from deep nesting |
| `MaxParseCollectionLen` | 16,777,216 | oversized list/map pre-allocation |
| `MaxParseByteLen` | 1,073,741,824 | oversized string/binary/ext allocation |
| `MaxDecimalExponent` | 1,048,576 | decimals that explode on later arithmetic |

## What is *not* canonical

- **JSON / `String()`** output is for humans. Its formatting is affected by
  package vars (`Base64Prefix`, `DecimalExpDelim`, `UnknownPrefix`); changing them
  changes that text but never the MessagePack bytes from `Pack`.
- **Mutability.** Only the `Immutable*` types are safe to share freely. The mutable
  types (`MutableList`/`SortedMap`/`SparseList`) may share backing arrays when
  `AllowFastAppends` is true (the default); treat a value derived from one as
  owning its parent's storage and do not mutate across goroutines. This does not
  affect the bytes of any single value, only aliasing.
