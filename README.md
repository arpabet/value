# value

![build workflow](https://go.arpabet.com/value/actions/workflows/build.yaml/badge.svg)

**Deterministic, canonical serialization for Go.** The same value always encodes
to exactly the same bytes, so those bytes can be **hashed and signed**
reproducibly. That is what makes `value` a natural fit for cryptographic
protocols and the **blockchain** space — content addressing, Merkle trees, and
signature verification all require a stable, canonical byte representation. The
exact wire format is specified in [CANONICAL.md](CANONICAL.md).

* **Deterministic & canonical:** equal values produce identical MessagePack
  bytes, so they hash identically — the foundation for content addressing,
  signing, and verification.
* **Immutable:** the `Immutable*` value types are safe to share across
  goroutines; mutable variants are available for in-place building.
* **Acyclic by construction:** values form trees, so packing never cycles.

## Install

```bash
go get go.arpabet.com/value
```

Requires Go 1.25+.

## Data types

Every value packs to canonical MessagePack via `value.Pack`; `value.Hex` renders
those bytes. All snippets below are compiled and output-checked in
[`example_test.go`](example_test.go).

### Null & Boolean

```go
value.Hex(value.Null)           // c0
value.Hex(value.Boolean(true))  // c3
value.Hex(value.Boolean(false)) // c2
```

### Numbers — `Long`, `Double`, `BigInt`, `Decimal`

```go
value.Hex(value.Long(123))               // 7b
value.Hex(value.Double(1.5))             // cb3ff8000000000000
value.Hex(value.BigInt(big.NewInt(123))) // d501027b  (msgpack extension)

value.Decimal(decimal.New(12345, -2)).Double() // 123.45  (exact base-10)
```

The four number kinds are distinct values with distinct encodings: `Long(1)`,
`Double(1.0)`, `BigInt(1)` and `Decimal(1)` are never `Equal`. See
[CANONICAL.md](CANONICAL.md) for the equality contract.

### Strings — `Utf8` (text) and `Raw` (binary)

```go
value.Hex(value.Utf8("hello"))               // a568656c6c6f  (msgpack str)
value.Hex(value.Raw([]byte{1, 2, 3}, false)) // c403010203    (msgpack bin)
```

### Lists — dense and sparse

A dense list encodes as a MessagePack array:

```go
l := value.EmptyMutableList()
l = l.Append(value.Long(1))
l = l.Append(value.Utf8("x"))

value.Hex(l)     // 9201a178
value.Jsonify(l) // [1,"x"]
```

A sparse list is keyed by index and encodes as an integer-keyed map:

```go
l := value.EmptySparseList()
l = l.PutAt(0, value.Long(10))
l = l.PutAt(5, value.Long(50))

value.Hex(l) // 82000a0532
```

### Maps

```go
m := value.EmptyImmutableMap()
m = m.Put("name", value.Utf8("ann"))
m = m.Put("age", value.Long(30))

value.Jsonify(m) // {"age": 30,"name": "ann"}   (keys sorted)
```

## Deterministic ordering

Canonical bytes depend only on contents, never on construction order — map keys
are always sorted before encoding. The same value built two different ways
produces identical bytes, and therefore the same hash:

```go
a := value.EmptyImmutableMap().Put("zebra", value.Long(1)).Put("apple", value.Long(2))
b := value.ImmutableMapOf(map[string]value.Value{
    "apple": value.Long(2),
    "zebra": value.Long(1),
})

value.Hex(a)                 // 82a56170706c6502a57a6562726101
value.Hex(a) == value.Hex(b) // true
```

## Pack / Unpack

```go
v := value.MutableList([]value.Value{
    value.Boolean(true), value.Long(123), value.Utf8("text"),
})

mp, _ := value.Pack(v) // canonical bytes
back, _ := value.Unpack(mp, false)

v.Equal(back) // true
```

## Hashing

Equal values hash identically across processes and machines — the basis for
content addressing and Merkle structures. `value.Hash` packs, then hashes:

```go
_, digest, _ := value.Hash(value.Long(123), crypto.SHA256)

hex.EncodeToString(digest)
// 021fb596db81e6d02bf3d2586ee3981fe519f275c0ac9ca76bbcf2ebb4097d96
// == sha256(0x7b), the canonical encoding of Long(123)
```

## Signing

Sign the canonical bytes directly. Verification is stable because the same value
always packs to the same bytes:

```go
pub, priv, _ := ed25519.GenerateKey(rand.Reader)

v := value.ImmutableMapOf(map[string]value.Value{
    "amount": value.Long(100),
    "to":     value.Utf8("alice"),
})
data, _ := value.Pack(v)

sig := ed25519.Sign(priv, data)
ed25519.Verify(pub, data, sig) // true
```

## Sealing (authenticated encryption)

`Seal`/`Unseal` wrap a value with NaCl box:

```go
recipientPub, recipientPriv, _ := box.GenerateKey(rand.Reader)
senderPub, senderPriv, _ := box.GenerateKey(rand.Reader)

sealed, _ := value.Seal(value.Utf8("secret message"), recipientPub, senderPriv)
out, _ := value.Unseal(sealed, senderPub, recipientPriv)

out.Equal(value.Utf8("secret message")) // true
```

## Documentation

- [CANONICAL.md](CANONICAL.md) — the exact canonical wire format and equality
  contract.
- [`example_test.go`](example_test.go) — every snippet above, compiled and
  output-checked by `go test`.

## License

[Apache-2.0](LICENSE).
