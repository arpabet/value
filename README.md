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

## Go struct schemas

Value deliberately provides two struct schema families. Both are deterministic,
but they represent different schema styles and produce different wire shapes.
Neither is a replacement for the other.

| API | Struct tags | Wire keys | Field types | Best suited for |
|---|---|---|---|---|
| `ValueCodec.Marshal` / `Unmarshal` | `value:"name"`, `json:"name"`, or `msgpack:"name"` | UTF-8 field names | ordinary Go types and `Value` | durable records, signed objects, APIs, and data that benefits from readable field names |
| `PackStruct` / `UnpackStruct` | `tag:"1"`, optionally `repeated:"true"` | integer field numbers | `Value`, slices of `Value`, and nested schema pointers | compact RPC payloads and binary protocol dialects with a separately defined schema |

The numeric codec is intentionally protobuf-style: field names do not appear on
the wire, and a stable integer identifies each field. It is not deprecated, it
is not protobuf wire-compatible, and it is not shorthand for
`Pack(Marshal(obj))`. Peers must agree on which codec defines a payload.

### Value codecs and tag dialects

`ValueCodec` converts an ordinary Go object to the same canonical `Value` model
while selecting which struct tags provide its field names:

```go
classic := value.DefaultValueCodec() // value:"field"
jsonDTO := value.JSONValueCodec()    // json:"field"
legacy := value.MsgpackValueCodec()  // msgpack:"field"

tree, _ := legacy.Marshal(record)
wire, _ := value.Pack(tree)
```

All codecs produce a `Value`; their names describe the struct-tag dialect, not
the final bytes. In particular, `JSONValueCodec` does not produce textual JSON,
and `MsgpackValueCodec` does not reproduce vmihailenco/msgpack bytes. `Pack`
always applies Value's canonical MessagePack encoding.

| Factory | Struct tag | Decode policy | Supported tag options |
|---|---|---|---|
| `DefaultValueCodec()` | `value` | fields required unless `omitempty` | `-`, `omitempty`, legacy signing selectors |
| `JSONValueCodec()` | `json` | missing fields allowed | `-`, `omitempty` |
| `MsgpackValueCodec()` | `msgpack` | missing fields allowed | `-`, `omitempty` |

Unsupported JSON/MessagePack options return an error instead of silently
claiming compatibility. Library-specific behaviors such as JSON `string` and
MessagePack `intern`, `inline`, or `as_array` are not emulated by these tag
profiles in this release. The dedicated `sign:"selector"` tag works with every
profile.

### Named map schema: `ValueCodec`

`Marshal` and `Unmarshal` map ordinary Go structs through `value` tags. The tag
defines the canonical schema; MessagePack is the byte encoding produced later
by `Pack`, not the struct-tag namespace.

The package-level functions are compatibility wrappers around
`DefaultValueCodec()`.

```go
type Record struct {
    ID   string `value:"id"`
    Note string `value:"note,omitempty"`
}

tree, _ := value.Marshal(Record{ID: "r1"})
wire, _ := value.Pack(tree)
```

Use `value` tags for stable protocol, storage, hashing, and signing field names.
Do not use dependency-injection configuration tags as serialization metadata.
An untagged exported field uses its Go field name; explicit tags are preferred
for durable or cryptographic schemas.

### Numeric binary schema: `PackStruct` / `UnpackStruct`

`PackStruct` writes canonical MessagePack directly from a numeric schema. It
sorts fields by number, so Go declaration order and Go field names do not affect
the bytes. `UnpackStruct` applies the same numeric schema while decoding.

```go
type RPCFrame struct {
    Method  value.String   `tag:"1"`
    Payload value.String   `tag:"2"`
    Trace   []value.String `tag:"3" repeated:"true"`
}

frame := &RPCFrame{
    Method:  value.Utf8("mail.deliver"),
    Payload: value.Raw(ciphertext, false),
    Trace:   []value.String{value.Utf8("edge"), value.Utf8("relay")},
}

wire, _ := value.PackStruct(frame)

var decoded RPCFrame
_ = value.UnpackStruct(wire, &decoded, false)
```

Numeric-schema rules:

* pass a pointer to `PackStruct` and `UnpackStruct` (apart from packing `nil`);
* give every field a unique, stable integer `tag`; never renumber or reuse a tag
  once a protocol has shipped;
* scalar fields must implement `value.Value`; nested messages are pointers to
  structs governed by the same numeric schema;
* a slice without `repeated:"true"` is encoded as one list-valued field;
* a slice with `repeated:"true"` is encoded as repeated occurrences of its
  numeric field key, matching the protobuf-style repeated-field model;
* nil fields and nil slices are absent from the wire; and
* unknown numeric tags are rejected, so adding a field requires coordinated
  schema/version negotiation rather than protobuf's unknown-field behavior.

Use `UnpackStruct` to decode numeric payloads, particularly repeated fields. A
generic `Unpack` produces a dynamic Value tree and does not carry the external
Go schema needed to reconstruct the typed message.

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

For Go structs, keep canonical field names and signing membership separate. The
preferred style is `value` for the wire name and `sign` for one or more signing
selectors:

```go
type License struct {
    Domain  string `value:"_dom" sign:"license"`
    Product string `value:"product" sign:"license,audit"`
    Sig     []byte `value:"sig"`
}

payload, _ := value.SignBytes(license, "license")
signature := ed25519.Sign(privateKey, payload)
```

The `sign` tag does not change ordinary `Marshal` or `Unmarshal`. Existing code
using selectors inside `value` tags, such as `value:"product,license"`, remains
supported for compatibility and produces the same projection, but new code
should use the dedicated `sign` tag.

## Sealing (authenticated encryption)

`Seal`/`Unseal` wrap a value with NaCl box:

```go
recipientPub, recipientPriv, _ := box.GenerateKey(rand.Reader)
senderPub, senderPriv, _ := box.GenerateKey(rand.Reader)

sealed, _ := value.Seal(value.Utf8("secret message"), recipientPub, senderPriv)
out, _ := value.Unseal(sealed, senderPub, recipientPriv)

out.Equal(value.Utf8("secret message")) // true
```

## How it compares: MessagePack `value` vs RLP vs CBOR

All three are binary encodings used where a **stable, canonical byte
representation** is needed for hashing and signing — but they make different
trade-offs.

| | `value` (this library) | RLP (Ethereum) | CBOR (RFC 8949) |
|---|---|---|---|
| Encoding | MessagePack | Recursive Length Prefix | CBOR |
| Data model | rich, dynamic — null, bool, int, float, big.Int, decimal, text, bytes, list, sparse list, map | two types only — byte strings and lists | rich, self-describing — ints, floats, bytes, text, arrays, maps, tags |
| Self-describing | yes | **no** (types come from an external schema) | yes |
| Deterministic output | **always** (canonical by design) | **always** (a single valid encoding) | **optional** (must opt into Core Deterministic Encoding, §4.2) |
| Native maps | yes, keys sorted | no (key/value lives in app-layer tries) | yes (sorted keys in deterministic mode) |
| Floats | float64 | none | float16/32/64 |
| Canonical spec | [CANONICAL.md](CANONICAL.md) | Ethereum Yellow Paper | IETF STD 94 |
| Cross-language standard | no (a Go library) | all Ethereum clients | yes, many languages |
| Typical use | hashing / signing / content addressing in Go | Ethereum tx, block & state hashing | COSE signing, CWT, WebAuthn, IoT |

**RLP** is the minimalist. It knows only byte strings and lists and has exactly
one canonical encoding, which is why Ethereum uses it to hash transactions,
blocks, and state. The cost is that it is **not self-describing**: the integer
`1024` and the two-byte string `{0x04, 0x00}` encode to the same bytes, and only
the surrounding schema decides which it is — so RLP is a poor general-purpose,
decode-anywhere value format.

**CBOR** is the standard. It is an IETF format with a rich, self-describing model,
tags for bignums/decimals/timestamps, and broad cross-language support, which is
why it underpins COSE signing, CWT tokens, and WebAuthn. Its catch for hashing is
that determinism is **opt-in**: plain CBOR permits several encodings of the same
value (map-key order, integer and float width), so you must explicitly select
Core Deterministic Encoding and a conforming library.

**`value`** sits between them for Go services. It offers a rich, self-describing
tree like CBOR, but is **canonical by default** like RLP — there is no
non-deterministic mode to forget to turn off — and it adds immutable structural
sharing plus one tree that emits both canonical bytes and JSON. The trade-offs:
it is a Go library rather than a multi-language standard, and its big-integer and
decimal extensions currently use a Go-specific framing (a portable framing is
planned for v2; see [CANONICAL.md](CANONICAL.md)).

## Documentation

- [CANONICAL.md](CANONICAL.md) — the exact canonical wire format and equality
  contract.
- [`example_test.go`](example_test.go) — every snippet above, compiled and
  output-checked by `go test`.

## License

[Apache-2.0](LICENSE).
