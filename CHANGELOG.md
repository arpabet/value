# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/), and the project aims to follow
[Semantic Versioning](https://semver.org/).

## v1.6.1 — 2026-11-11

Unsigned integers above the range of a long keep their number.

### Fixed

- **`Marshal` of an unsigned field above `math.MaxInt64`** (`uint64`, `uint`)
  wrote the negative long of the same bits: `uint64(1<<63)` went out as
  `-9223372036854775808`, which another reader takes for a negative number and
  `Unmarshal` with `Lossless()` refused ("overflows uint64"). The field is now
  written as the big integer it is, in the encoding
  [CANONICAL.md](CANONICAL.md#integers-long) already gives that number
  (`c70901008000000000000000`). A value within the range of a long is written
  as before.
- **`Unpack` and `Read` of MessagePack's `uint 64` above the range of a long**
  (`cf` with the top bit set), which a canonical writer never produces and
  other MessagePack writers do, gave the negative long of the same bits. They
  give the big integer of that number, which packs again canonically.
- `ParseNumber` of an integer text beyond the range of a long
  (`"18446744073709551615"`) gave the nearest double and gives the big integer
  now; `GetNumber` and `GetNumberAt` on a string value use it.
- `Number.BigInt()` of a decimal gave the coefficient of its floor: 1 for
  `1e3`, and -2 for -1.5 where `Long()` gives -1. It gives the integer part,
  rounded toward zero as `Long()` rounds.

### Added

- `Unsigned(uint64) Number`: the number of an unsigned integer, a long while it
  fits one and a big integer above. `Long(int64(u))` is a negative number for
  such a `u`.

### Compatibility

- **Stored data.** An unsigned field above the range of a long that an earlier
  version wrote is a negative long. The default codec reads it back into an
  unsigned field as before, since it wraps; the lossless codec refuses it, as
  it always did. Reading with the default codec and writing again migrates it.
- **Signatures and hashes.** The bytes of a value with such a field change, and
  with them its `SignBytes`, `SignHash` and `Hash`: a signature an earlier
  version made over one does not verify against the projection this version
  builds. Fields within the range of a long are not affected.
- **Earlier readers** read the new encoding: the lossless codec exactly, the
  default codec through `Number.Long()` of the big integer, which is its low 64
  bits.

## v1.6.0 — 2026-11-10

Lossless JSON.

### Added

- **`ParseJSON`, `FormatJSON` and `AppendJSON`** with two modes, `JSONCanonical`
  and `JSONRelaxed` ([CANONICAL.md](CANONICAL.md#json)). Kinds plain JSON lacks
  are written as single-key wrapper objects (`$long`, `$double`, `$bigint`,
  `$decimal`, `$bin`, `$timestamp`, `$sparse`, `$ext`, and `$map` for a map whose
  only key is a wrapper key); canonical mode wraps every one of them, so a round
  trip keeps the kinds and exact digits, and relaxed mode writes numbers as plain
  JSON numbers. `ParseJSON` reads a bare JSON number as a `Decimal` from its
  exact text, integers included, so an untyped number never changes kind with
  its size.
- `DoubleText` and `DecimalText`: the exact texts the JSON modes use.
- `FuzzParseJSON`.
- **Lossless number conversion** (`ValueCodec.Lossless()`): `Unmarshal` sets a
  numeric field only when the number survives the conversion and fails naming
  the field otherwise ([CANONICAL.md](CANONICAL.md#numbers-into-go-fields));
  `ExactInt`, `ExactUint`, `ExactFloat`, `ExactDecimal`, `ExactBigInt`.
- `ValueCodec.AllowMissing()`: absent fields stay at their zero value, for
  readers of structs that gained fields.
- `ValueCodec.StructFields`: the fields of a struct type as the codec maps them.
- An interface field, such as `any`, unmarshals the value's tree (a field of
  the interface `Value` already did); it used to fail.
- **`decimal.Decimal` and `big.Int` fields** marshal as `Decimal` and `BigInt`
  and unmarshal from any number that converts exactly. They used to marshal as
  empty maps.
- `jsonCanonical` in `testdata/canonical.json`: the canonical lossless JSON of
  every vector.

### Unchanged

- `Jsonify`, `PrintJSON` and `MarshalJSON` keep their rendering (the `base64,`
  prefix, hexadecimal big numbers), which value-rpc's JSON wire codec speaks.

## v1.5.0 — 2026-11-03

Timestamps and portable big numbers. **The wire format of `BigInt` and
`Decimal` changes**; see _Compatibility_ before upgrading a system that stores
or exchanges those two kinds.

### Added

- **`TIMESTAMP` kind.** `value.Timestamp` is an instant with nanosecond
  precision and no location, built with `Time(t time.Time)`,
  `UnixTime(sec, nsec)` or `ParseTimestamp(rfc3339)`, read with `Time()` and
  `Unix()`. It encodes as MessagePack's predefined timestamp extension (type
  -1, `TimestampExt`) in the smallest of its three forms, decodes from any of
  them (`UnpackTimestamp`), compares by instant, and renders as RFC 3339 in UTC
  in `String()` and JSON. `Kind.String()` now also names `NULL`.
- **`time.Time` support in every codec profile.** `Marshal` (the `value`,
  `json` and `msgpack` tag profiles alike) writes a `time.Time` field as a
  `Timestamp`; `Unmarshal` reads a `Timestamp` and, for data written by earlier
  versions, an integer of Unix milliseconds. `PackStruct`/`UnpackStruct` carry
  `value.Timestamp` fields like any other `Value`.
- **Conformance vectors.** `testdata/canonical.json` describes 80 values in a
  language-neutral notation with their canonical bytes; `vectors_test.go` runs
  them. [CANONICAL.md](CANONICAL.md) is now normative for every implementation.
- `LegacyBigNumbers`, `UnpackLegacyBigInt` and `UnpackLegacyDecimal` for reading
  data written before this release (below).

### Changed

- **`BigInt` (ext 1) payload** is the two's-complement big-endian integer in
  its minimal length (`BigInteger.toByteArray()`): `BigInt(123)` is now
  `d4017b`, previously `d501027b` (Go's gob framing).
- **`Decimal` (ext 2) payload** is a 4-byte big-endian signed scale followed
  by the two's-complement unscaled integer, `unscaled × 10^-scale`
  (`BigDecimal(unscaled, scale)`, Avro's layout): `Decimal(1.10)` is now
  `c70502000000026e`. `MaxDecimalExponent` bounds the absolute scale.
- Decoding accepts longer, sign-extended two's-complement encodings and
  re-encodes them minimally, so `Equal` values still pack identically.
- `time.Time` marshals as a `Timestamp` instead of an integer of Unix
  milliseconds, keeping nanoseconds.

### Compatibility

- A reader older than v1.5.0 cannot read `BigInt`, `Decimal` or `time.Time`
  values written by this version (it reads the first two as different numbers
  and the third as an unknown extension). Upgrade readers before writers.
- This version reads `Long`, `Double`, strings, binaries, lists and maps from
  any earlier version unchanged, and a `time.Time` stored as Unix milliseconds.
  It does **not** read the old `BigInt` and `Decimal` payloads by default: the
  two layouts cannot be told apart from the bytes, so there is no guessing. A
  program with stored big numbers sets `value.LegacyBigNumbers = true` while it
  migrates (decode with the old layout, `Pack` writes the new one) or calls
  `UnpackLegacyBigInt`/`UnpackLegacyDecimal` itself.
- `Kind` gains `TIMESTAMP` after `UNKNOWN`; the other constants keep their
  values. Code that switches over `Kind` exhaustively needs a new case.
- The `Packer`, `Unpacker` and `Parser` interfaces are unchanged.

## v1.4.0 — 2026-08-15

### Added

- Added the immutable `ValueCodec` abstraction with
  `DefaultValueCodec()`, `JSONValueCodec()`, and `MsgpackValueCodec()` tag
  profiles. Every profile maps objects to the canonical `Value` tree; byte
  encoding remains the separate responsibility of `Pack`/`Unpack`.
- Added the dedicated `sign:"selector"` struct tag for `SignBytes` and
  `SignHash`. Multiple comma-separated selectors are supported.
- Documented the preferred separation between canonical field naming
  (`value:"field"`) and signing policy (`sign:"license"`).
- Documented `PackStruct` / `UnpackStruct` as the supported protobuf-style
  numeric binary dialect for RPC and other externally schematized protocols,
  distinct from the named-map `Marshal` / `Unmarshal` codec.

### Compatibility

- Package-level `Marshal`, `Unmarshal`, `SignBytes`, and `SignHash` remain
  wrappers over the classic `value`-tag profile. The zero `ValueCodec` also
  selects that profile.
- JSON and MessagePack profiles support field names, `-`, and `omitempty`.
  Unsupported library-specific tag options are rejected explicitly, and these
  profiles do not claim byte compatibility with encoding/json or third-party
  MessagePack encoders.
- Existing signing selectors embedded in `value` tags, such as
  `value:"field,license"`, remain supported and produce the same canonical
  projection. Dedicated and legacy selectors may coexist on the same field.
- The canonical MessagePack wire format and ordinary `Marshal`/`Unmarshal`
  behavior are unchanged.

## v1.2.0 — 2026-06-13

First openly-licensed, hardened, and fully-specified release. The wire format is
unchanged; see _Behavior changes_ before upgrading.

### License

- **Relicensed from BUSL-1.1 to [Apache-2.0](LICENSE).** The library is now
  permissively licensed (OSI-approved, with a patent grant). Copyright remains
  © Karagatan LLC.

### Added

- **Canonical-form specification** ([CANONICAL.md](CANONICAL.md)) — the exact
  MessagePack encoding, the equality contract, and the decoding limits, backed by
  golden-vector tests. Determinism is now the documented headline feature: equal
  values produce identical bytes, enabling hashing, signing, content addressing,
  and Merkle structures.
- **Configurable decoding limits** (`limits.go`): `MaxParseDepth`,
  `MaxParseCollectionLen`, `MaxParseByteLen`, `MaxDecimalExponent` (set any to 0
  to disable).
- **Fuzz targets** `FuzzUnpack` and `FuzzRead`, plus a CI fuzzing job and a
  `make fuzz` target.
- **Generics & iterators** (`iter.go`): `As[T]` for typed extraction, and
  `MapAll`/`ListAll` returning `iter.Seq2` for range-over-func.
- **Real benchmarks** `BenchmarkPack` / `BenchmarkUnpack`.
- **Runnable examples** ([`example_test.go`](example_test.go)) for every data
  type plus hashing, signing (ed25519), and sealing (NaCl box). They render on
  pkg.go.dev and are output-checked by `go test`.
- **`ApproxEqual(a, b Number)`** providing the previous tolerant, cross-type
  numeric comparison.

### Changed

- **Requires Go 1.25** (was 1.17). CI now runs a Go 1.25/1.26 matrix with
  `go vet`, the race detector, `govulncheck`, and fuzzing.
- **Dependencies upgraded**: `shopspring/decimal` v1.4.0, `stretchr/testify`
  v1.11.1, `golang.org/x/crypto` v0.53.0. Removed the archived
  `github.com/pkg/errors` (migrated to stdlib `fmt.Errorf`).
- **README** rewritten with correct, verified examples (the previous examples
  referenced API that did not exist).

### Fixed

- **Large lists were corrupted**: arrays with more than 65535 elements emitted
  the `array16` type byte with a 32-bit length, desynchronizing the stream. Now
  emit `array32`.
- **Write errors were silently dropped**: the packer used value receivers, so
  `io.Writer` failures never surfaced from `Pack`/`Write`. Errors now propagate.
- **Mixed integer/string-keyed maps** could panic or lose entries while decoding
  (wrong back-fill index and a missing re-sort). Fixed.
- **`Number.Equal` was asymmetric and inconsistent with serialization**
  (`Long(1).Equal(Double(1.5))` was true while the reverse was false). It is now
  symmetric, reflexive (including NaN), and consistent with the packed bytes.
- **Stream-reader robustness**: `Read`/`ReadStream` panicked on readers that are
  not `io.ByteReader` (e.g. `os.File`, `net.Conn`) and mis-framed values on short
  reads. They now use a comma-ok assertion and `io.ReadFull`.
- **`Unseal` panicked** on inputs shorter than the 24-byte nonce; it now returns
  an error.
- **`ReadStream`** returned an error on a clean end-of-stream and contained
  unreachable code; it now returns `nil` on `io.EOF`.

### Security

The decoder is now hardened against hostile or corrupt input:

- list/map allocation is incremental and bounded, so a tiny header claiming
  billions of elements can no longer exhaust memory;
- string, binary, extension sizes and nesting depth are bounded;
- a decimal with an enormous exponent is rejected at decode (it could otherwise
  make `IntPart`/arithmetic materialize a giant integer — this was found by
  fuzzing);
- sparse-list indices are restricted to a sane range so `Len()`/`Values()`
  cannot be coerced into allocating a huge slice.

### Behavior changes (review before upgrading)

- `Number.Equal` no longer coerces across number types or tolerates floating
  differences: `Long(1)`, `Double(1.0)`, `BigInt(1)` and `Decimal(1)` are all
  distinct, and two doubles are equal only if their bits match. Use `ApproxEqual`
  for the previous behavior.
- The decoder rejects inputs that exceed the new limits and decimals with
  out-of-range exponents. Tune the `MaxParse*` / `MaxDecimalExponent` package
  variables (0 disables a check) if you need different bounds.
- A map whose keys are large or out-of-range integers now decodes as a
  string-keyed map rather than a sparse list.

### Deferred to a future major version (v2)

The wire format is intentionally unchanged in this release. The following are
documented in [CANONICAL.md](CANONICAL.md) as breaking changes for v2: a
portable, language-neutral encoding for big integers and decimals (currently Go
`gob` / `decimal.MarshalBinary`); removing the global formatting knobs
(`Base64Prefix`, `DecimalExpDelim`, `UnknownPrefix`); and enforcing immutability
on the mutable collection types.
