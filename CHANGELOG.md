# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/), and the project aims to follow
[Semantic Versioning](https://semver.org/).

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
