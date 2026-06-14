/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value_test

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	_ "crypto/sha256" // register SHA-256 for crypto.SHA256
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/shopspring/decimal"
	val "go.arpabet.com/value"
	"golang.org/x/crypto/nacl/box"
)

func ExampleNull() {
	fmt.Println(val.Hex(val.Null))
	// Output: c0
}

func ExampleBoolean() {
	fmt.Println(val.Hex(val.Boolean(true)), val.Hex(val.Boolean(false)))
	// Output: c3 c2
}

func ExampleLong() {
	fmt.Println(val.Hex(val.Long(123)))
	// Output: 7b
}

func ExampleDouble() {
	fmt.Println(val.Hex(val.Double(1.5)))
	// Output: cb3ff8000000000000
}

func ExampleBigInt() {
	fmt.Println(val.Hex(val.BigInt(big.NewInt(123))))
	// Output: d501027b
}

func ExampleDecimal() {
	d := val.Decimal(decimal.New(12345, -2)) // 123.45, exact base-10
	fmt.Println(d.Double())
	// Output: 123.45
}

func ExampleUtf8() {
	fmt.Println(val.Hex(val.Utf8("hello")))
	// Output: a568656c6c6f
}

func ExampleRaw() {
	fmt.Println(val.Hex(val.Raw([]byte{1, 2, 3}, false)))
	// Output: c403010203
}

func ExampleMutableList() {
	l := val.EmptyMutableList()
	l = l.Append(val.Long(1))
	l = l.Append(val.Utf8("x"))
	fmt.Println(val.Hex(l), val.Jsonify(l))
	// Output: 9201a178 [1,"x"]
}

func ExampleSparseList() {
	// A sparse list is keyed by index and encodes as an integer-keyed map.
	l := val.EmptySparseList()
	l = l.PutAt(0, val.Long(10))
	l = l.PutAt(5, val.Long(50))
	fmt.Println(val.Hex(l))
	// Output: 82000a0532
}

func ExampleImmutableMap() {
	m := val.EmptyImmutableMap()
	m = m.Put("name", val.Utf8("ann"))
	m = m.Put("age", val.Long(30))
	fmt.Println(val.Jsonify(m))
	// Output: {"age": 30,"name": "ann"}
}

// Determinism: the canonical bytes depend only on contents, not on how the
// value was built — map keys are always sorted.
func Example_determinism() {
	a := val.EmptyImmutableMap().Put("zebra", val.Long(1)).Put("apple", val.Long(2))
	b := val.ImmutableMapOf(map[string]val.Value{"apple": val.Long(2), "zebra": val.Long(1)})
	fmt.Println(val.Hex(a))
	fmt.Println(val.Hex(a) == val.Hex(b))
	// Output:
	// 82a56170706c6502a57a6562726101
	// true
}

func Example_packUnpack() {
	v := val.MutableList([]val.Value{val.Boolean(true), val.Long(123), val.Utf8("text")})
	mp, _ := val.Pack(v)
	back, _ := val.Unpack(mp, false)
	fmt.Println(v.Equal(back))
	// Output: true
}

// Hashing: equal values hash identically, across processes and machines.
func Example_hashing() {
	v := val.Long(123)
	_, digest, _ := val.Hash(v, crypto.SHA256)
	fmt.Println(hex.EncodeToString(digest))
	// Output: 021fb596db81e6d02bf3d2586ee3981fe519f275c0ac9ca76bbcf2ebb4097d96
}

// Signing: sign the canonical bytes. Verification is stable because the same
// value always packs to the same bytes.
func Example_signing() {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)

	v := val.ImmutableMapOf(map[string]val.Value{"amount": val.Long(100), "to": val.Utf8("alice")})
	data, _ := val.Pack(v)

	sig := ed25519.Sign(priv, data)
	fmt.Println(ed25519.Verify(pub, data, sig))
	// Output: true
}

// Seal/Unseal: authenticated encryption of a value with NaCl box.
func Example_seal() {
	recipientPub, recipientPriv, _ := box.GenerateKey(rand.Reader)
	senderPub, senderPriv, _ := box.GenerateKey(rand.Reader)

	v := val.Utf8("secret message")
	sealed, _ := val.Seal(v, recipientPub, senderPriv)
	out, _ := val.Unseal(sealed, senderPub, recipientPriv)

	fmt.Println(out.Equal(v))
	// Output: true
}
