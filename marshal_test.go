/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value

import (
	"bytes"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"reflect"
	"testing"
	"time"
)

type inner struct {
	Epoch uint64 `value:"epoch"`
	Root  []byte `value:"root"`
}

// roundTripMsg covers the Go field shapes depecher's wire types use.
type roundTripMsg struct {
	Account string            `value:"acc"`
	Device  string            `value:"dev"`
	Count   uint32            `value:"count"`
	Seq     int64             `value:"seq"`
	Live    bool              `value:"live"`
	DhPub   []byte            `value:"dh"`
	OPKs    map[string][]byte `value:"opk"`
	Names   []string          `value:"names"`
	Bundle  inner             `value:"bundle"`
	Skipped string            `value:"-"`
	private int               // unexported, ignored
}

func TestMarshalRoundTrip(t *testing.T) {
	src := roundTripMsg{
		Account: "vox-abc", Device: "dev-1", Count: 42, Seq: -7, Live: true,
		DhPub: []byte{1, 2, 3, 4}, OPKs: map[string][]byte{"k1": {9, 9}, "k2": {8}},
		Names:   []string{"alice", "bob"},
		Bundle:  inner{Epoch: 100, Root: []byte{0xaa, 0xbb}},
		Skipped: "should not survive", private: 5,
	}

	v, err := Marshal(src)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if v.Kind() != MAP {
		t.Fatalf("expected MAP, got %s", v.Kind())
	}
	if got := v.(Map).Get("Skipped"); got != nil && got.Kind() != NULL {
		t.Fatalf(`value:"-" field leaked into the map`)
	}

	var dst roundTripMsg
	if err := Unmarshal(v, &dst); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// Skipped/private don't round-trip; compare the rest.
	src.Skipped, src.private = "", 0
	if !reflect.DeepEqual(src, dst) {
		t.Fatalf("round-trip mismatch:\n src=%+v\n dst=%+v", src, dst)
	}

	// Through the wire bytes too (Pack/Unpack).
	packed, err := Pack(v)
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	uv, err := Unpack(packed, true)
	if err != nil {
		t.Fatalf("unpack: %v", err)
	}
	var dst2 roundTripMsg
	if err := Unmarshal(uv, &dst2); err != nil {
		t.Fatalf("unmarshal wire: %v", err)
	}
	if !reflect.DeepEqual(src, dst2) {
		t.Fatalf("wire round-trip mismatch:\n src=%+v\n dst=%+v", src, dst2)
	}
}

// TestMarshalCanonical: marshaling the same struct (whose map field iterates in
// random Go order) must pack to identical, canonical bytes — the property that
// makes signing stable.
func TestMarshalCanonical(t *testing.T) {
	m := roundTripMsg{
		Account: "vox-abc",
		OPKs:    map[string][]byte{"a": {1}, "b": {2}, "c": {3}, "d": {4}, "e": {5}},
	}
	first, err := SignBytesOrPack(t, m)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		again, err := SignBytesOrPack(t, m)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(first, again) {
			t.Fatalf("non-deterministic packing on iteration %d", i)
		}
	}
}

func SignBytesOrPack(t *testing.T, obj interface{}) ([]byte, error) {
	t.Helper()
	v, err := Marshal(obj)
	if err != nil {
		return nil, err
	}
	return Pack(v)
}

type signMsg struct {
	Domain string `value:"_dom,sign"` // domain separation
	Acct   string `value:"acc,sign"`
	Device string `value:"dev,sign"`
	DhPub  []byte `value:"dh,sign"`
	Sig    []byte `value:"sig"`   // the signature itself — NOT signed
	Nonce  uint64 `value:"nonce"` // metadata — NOT signed
}

func TestSignProjection(t *testing.T) {
	base := signMsg{Domain: "dr_ack/v1", Acct: "vox-a", Device: "dev-1", DhPub: []byte{1, 2, 3}}

	// Differ only in non-sign fields -> identical signing bytes.
	a := base
	a.Sig = []byte{0xde, 0xad}
	a.Nonce = 1
	b := base
	b.Sig = []byte{0xbe, 0xef}
	b.Nonce = 999

	sa, err := SignBytes(a)
	if err != nil {
		t.Fatal(err)
	}
	sb, err := SignBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(sa, sb) {
		t.Fatal("non-sign fields affected the signing projection")
	}

	// Changing a sign field -> different signing bytes.
	c := base
	c.Acct = "vox-evil"
	sc, err := SignBytes(c)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(sa, sc) {
		t.Fatal("changing a signed field did not change the signing projection")
	}

	// End-to-end ed25519: the verifier rebuilds the struct from the wire and
	// re-derives the same projection.
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	sig := ed25519.Sign(priv, sa)

	wire, err := Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	packed, _ := Pack(wire)
	uv, _ := Unpack(packed, true)
	var recv signMsg
	if err := Unmarshal(uv, &recv); err != nil {
		t.Fatal(err)
	}
	verifyBytes, err := SignBytes(recv)
	if err != nil {
		t.Fatal(err)
	}
	if !ed25519.Verify(pub, verifyBytes, sig) {
		t.Fatal("signature did not verify after wire round-trip")
	}

	// SignHash is SignBytes + hash.
	h, err := SignHash(a, crypto.SHA256)
	if err != nil {
		t.Fatal(err)
	}
	if len(h) != 32 {
		t.Fatalf("SignHash SHA256 len = %d, want 32", len(h))
	}
}

// A field already of Value type passes through unchanged.
func TestMarshalValuePassthrough(t *testing.T) {
	type holder struct {
		Raw Value `value:"raw"`
	}
	h := holder{Raw: Long(123)}
	v, err := Marshal(h)
	if err != nil {
		t.Fatal(err)
	}
	if got := v.(Map).Get("raw"); got == nil || got.Kind() != NUMBER || got.(Number).Long() != 123 {
		t.Fatalf("Value passthrough failed: %v", got)
	}
}

func TestMarshalTime(t *testing.T) {
	type withTime struct {
		Name string    `value:"name"`
		At   time.Time `value:"at"`
	}
	src := withTime{Name: "x", At: time.Now()}
	v, err := Marshal(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst withTime
	if err := Unmarshal(v, &dst); err != nil {
		t.Fatal(err)
	}
	if !src.At.Truncate(time.Millisecond).Equal(dst.At) {
		t.Fatalf("time round-trip: %v != %v", src.At, dst.At)
	}
}
