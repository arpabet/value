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

	sa, err := SignBytes(a, "sign")
	if err != nil {
		t.Fatal(err)
	}
	sb, err := SignBytes(b, "sign")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(sa, sb) {
		t.Fatal("non-sign fields affected the signing projection")
	}

	// Changing a sign field -> different signing bytes.
	c := base
	c.Acct = "vox-evil"
	sc, err := SignBytes(c, "sign")
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
	verifyBytes, err := SignBytes(recv, "sign")
	if err != nil {
		t.Fatal(err)
	}
	if !ed25519.Verify(pub, verifyBytes, sig) {
		t.Fatal("signature did not verify after wire round-trip")
	}

	// SignHash is SignBytes + hash.
	h, err := SignHash(a, crypto.SHA256, "sign")
	if err != nil {
		t.Fatal(err)
	}
	if len(h) != 32 {
		t.Fatalf("SignHash SHA256 len = %d, want 32", len(h))
	}
}

func TestSignProjectionDedicatedTag(t *testing.T) {
	type preferred struct {
		Domain  string `value:"_dom" sign:"license"`
		Product string `value:"product" sign:"license,audit"`
		Meta    string `value:"meta"`
	}
	type legacy struct {
		Domain  string `value:"_dom,license"`
		Product string `value:"product,license,audit"`
		Meta    string `value:"meta"`
	}

	p := preferred{Domain: "license/v1", Product: "mailnite", Meta: "not signed"}
	l := legacy{Domain: p.Domain, Product: p.Product, Meta: p.Meta}

	preferredBytes, err := SignBytes(p, "license")
	if err != nil {
		t.Fatal(err)
	}
	legacyBytes, err := SignBytes(l, "license")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(preferredBytes, legacyBytes) {
		t.Fatal("dedicated and legacy signing tags produced different projections")
	}

	// Signing metadata does not affect the ordinary wire schema.
	wire, err := Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := wire.(Map).Get("meta"); got == nil || got.Kind() == NULL {
		t.Fatal("ordinary marshal omitted a field without a sign selector")
	}

	// The second selector includes only Product, proving comma-separated
	// dedicated selectors are independent.
	auditBytes, err := SignBytes(p, "audit")
	if err != nil {
		t.Fatal(err)
	}
	auditValue, err := Unpack(auditBytes, true)
	if err != nil {
		t.Fatal(err)
	}
	auditMap := auditValue.(Map).HashMap()
	if len(auditMap) != 1 || auditMap["product"] == nil {
		t.Fatalf("audit projection = %v, want only product", auditMap)
	}
}

func TestSignProjectionCombinesDedicatedAndLegacySelectors(t *testing.T) {
	type mixed struct {
		Both string `value:"both,legacy" sign:"modern"`
	}
	m := mixed{Both: "x"}
	legacyBytes, err := SignBytes(m, "legacy")
	if err != nil {
		t.Fatal(err)
	}
	modernBytes, err := SignBytes(m, "modern")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(legacyBytes, modernBytes) {
		t.Fatal("legacy and dedicated selectors must compose on the same field")
	}
}

// A struct can define several independent signatures via per-field markers; the
// SignBytes selector picks which one to project.
func TestSignMultiMarker(t *testing.T) {
	type multi struct {
		Dom1 string `value:"_d1,sig1"`       // sig1 only
		Dom2 string `value:"_d2,sig2"`       // sig2 only
		A    string `value:"a,sig1"`         // sig1 only
		B    string `value:"b,sig2"`         // sig2 only
		Both string `value:"both,sig1,sig2"` // in both signatures
		Meta string `value:"meta"`           // in neither
	}
	m := multi{Dom1: "d1", Dom2: "d2", A: "aval", B: "bval", Both: "x", Meta: "ignored"}

	s1, err := SignBytes(m, "sig1")
	if err != nil {
		t.Fatal(err)
	}
	s2, err := SignBytes(m, "sig2")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(s1, s2) {
		t.Fatal("different markers must produce different projections")
	}

	// A sig2-only field must NOT affect the sig1 projection...
	m2 := m
	m2.B = "changed"
	if s1b, _ := SignBytes(m2, "sig1"); !bytes.Equal(s1, s1b) {
		t.Fatal("a sig2-only field leaked into the sig1 projection")
	}
	// ...but a shared (sig1,sig2) field must affect sig1.
	m3 := m
	m3.Both = "changed"
	if s1c, _ := SignBytes(m3, "sig1"); bytes.Equal(s1, s1c) {
		t.Fatal("a shared field did not affect the sig1 projection")
	}
	// A sig1-only field must NOT affect the sig2 projection.
	mA := m
	mA.A = "changed"
	if s2b, _ := SignBytes(mA, "sig2"); !bytes.Equal(s2, s2b) {
		t.Fatal("a sig1-only field leaked into the sig2 projection")
	}

	// An empty marker is rejected — no silently-empty signature.
	if _, err := SignBytes(m, ""); err == nil {
		t.Fatal("empty selector must error")
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
