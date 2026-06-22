/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value

import "testing"

func TestMarshalOmitEmpty(t *testing.T) {
	type rec struct {
		ID   string `value:"id"`             // required
		Note string `value:"note,omitempty"` // optional
		Data []byte `value:"data,omitempty"` // optional
	}

	// Empty optional fields are dropped; the required field is always written.
	v, err := Marshal(rec{ID: "x"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	hm := v.(Map).HashMap()
	if _, ok := hm["id"]; !ok {
		t.Fatal("required field id missing")
	}
	if _, ok := hm["note"]; ok {
		t.Fatal("empty optional note should be omitted")
	}
	if _, ok := hm["data"]; ok {
		t.Fatal("empty optional data should be omitted")
	}

	// Non-empty optional fields are included.
	hm2 := mustMarshal(t, rec{ID: "x", Note: "hi", Data: []byte{1}}).(Map).HashMap()
	if _, ok := hm2["note"]; !ok {
		t.Fatal("non-empty optional note should be present")
	}
	if _, ok := hm2["data"]; !ok {
		t.Fatal("non-empty optional data should be present")
	}
}

func TestUnmarshalRequired(t *testing.T) {
	type rec struct {
		ID   string `value:"id"`
		Note string `value:"note,omitempty"`
	}

	// Missing required field -> error.
	missing := ImmutableMapOf(map[string]Value{"note": Utf8("hi")})
	var got rec
	if err := Unmarshal(missing, &got); err == nil {
		t.Fatal("expected error for missing required field id")
	}

	// Missing optional field -> ok, zero value.
	onlyReq := ImmutableMapOf(map[string]Value{"id": Utf8("x")})
	var got2 rec
	if err := Unmarshal(onlyReq, &got2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got2.ID != "x" || got2.Note != "" {
		t.Fatalf("bad decode: %+v", got2)
	}

	// Round-trip with an empty optional field: marshal omits it, unmarshal
	// tolerates the absence.
	var got3 rec
	if err := Unmarshal(mustMarshal(t, rec{ID: "y"}), &got3); err != nil {
		t.Fatalf("round-trip: %v", err)
	}
	if got3.ID != "y" {
		t.Fatalf("bad round-trip: %+v", got3)
	}
}

// TestSignIgnoresOmitEmpty: the signing projection includes every sign-tagged
// field even when empty + omitempty, so the signed bytes keep a fixed shape.
func TestSignIgnoresOmitEmpty(t *testing.T) {
	type sig struct {
		Domain string `value:"dom,sign"`
		Data   []byte `value:"data,sign,omitempty"`
	}
	v, err := signProjection(sig{Domain: "d"}) // Data empty
	if err != nil {
		t.Fatalf("signProjection: %v", err)
	}
	if _, present := v.(Map).HashMap()["data"]; !present {
		t.Fatal("sign projection must include an empty omitempty field (fixed signed shape)")
	}
}

func mustMarshal(t *testing.T, obj interface{}) Value {
	t.Helper()
	v, err := Marshal(obj)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return v
}
