/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value

import (
	"bytes"
	"testing"
)

type codecRecord struct {
	Name   string `value:"value_name" json:"json_name" msgpack:"msgpack_name" sign:"record"`
	Count  int    `value:"value_count,omitempty" json:"json_count,omitempty" msgpack:"msgpack_count,omitempty"`
	Hidden string `value:"-" json:"-" msgpack:"-"`
}

func TestValueCodecProfilesSelectIndependentTags(t *testing.T) {
	tests := []struct {
		name      string
		codec     ValueCodec
		fieldName string
	}{
		{name: "value", codec: DefaultValueCodec(), fieldName: "value_name"},
		{name: "json", codec: JSONValueCodec(), fieldName: "json_name"},
		{name: "msgpack", codec: MsgpackValueCodec(), fieldName: "msgpack_name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := tt.codec.Marshal(codecRecord{Name: "ann", Hidden: "secret"})
			if err != nil {
				t.Fatal(err)
			}
			m := v.(Map).HashMap()
			if len(m) != 1 || m[tt.fieldName] == nil {
				t.Fatalf("map = %v, want only %q", m, tt.fieldName)
			}

			var decoded codecRecord
			if err := tt.codec.Unmarshal(v, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded.Name != "ann" || decoded.Count != 0 || decoded.Hidden != "" {
				t.Fatalf("decoded = %+v", decoded)
			}
		})
	}
}

func TestValueCodecMissingFieldPolicies(t *testing.T) {
	empty := EmptyImmutableMap()
	var dst codecRecord
	if err := DefaultValueCodec().Unmarshal(empty, &dst); err == nil {
		t.Fatal("default codec accepted a missing required field")
	}
	if err := JSONValueCodec().Unmarshal(empty, &dst); err != nil {
		t.Fatalf("JSON codec rejected a missing field: %v", err)
	}
	if err := MsgpackValueCodec().Unmarshal(empty, &dst); err != nil {
		t.Fatalf("MessagePack codec rejected a missing field: %v", err)
	}
}

func TestZeroValueCodecIsDefaultCodec(t *testing.T) {
	record := codecRecord{Name: "ann", Count: 7}
	zero, err := (ValueCodec{}).Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	zeroBytes, _ := Pack(zero)
	legacyBytes, _ := Pack(legacy)
	if !bytes.Equal(zeroBytes, legacyBytes) {
		t.Fatal("zero ValueCodec changed the default canonical representation")
	}
}

func TestValueCodecSigningUsesActiveFieldNames(t *testing.T) {
	record := codecRecord{Name: "ann", Count: 7}
	jsonBytes, err := JSONValueCodec().SignBytes(record, "record")
	if err != nil {
		t.Fatal(err)
	}
	v, err := Unpack(jsonBytes, true)
	if err != nil {
		t.Fatal(err)
	}
	m := v.(Map).HashMap()
	if len(m) != 1 || m["json_name"] == nil {
		t.Fatalf("JSON signing projection = %v, want json_name", m)
	}
}

func TestSignTagSupportsMultipleSelectors(t *testing.T) {
	type multiSignatureRecord struct {
		Domain1 string `value:"domain1" sign:"license1"`
		Domain2 string `value:"domain2" sign:"license2"`
		Product string `value:"product" sign:"license1,license2"`
		Owner   string `value:"owner" sign:"license1"`
		Plan    string `value:"plan" sign:"license2"`
		Note    string `value:"note"`
	}

	record := multiSignatureRecord{
		Domain1: "license/owner/v1",
		Domain2: "license/plan/v1",
		Product: "mailnite",
		Owner:   "example.com",
		Plan:    "enterprise",
		Note:    "not signed",
	}

	license1Bytes, err := SignBytes(record, "license1")
	if err != nil {
		t.Fatal(err)
	}
	license2Bytes, err := SignBytes(record, "license2")
	if err != nil {
		t.Fatal(err)
	}

	license1Value, err := Unpack(license1Bytes, true)
	if err != nil {
		t.Fatal(err)
	}
	license2Value, err := Unpack(license2Bytes, true)
	if err != nil {
		t.Fatal(err)
	}

	license1 := license1Value.(Map).HashMap()
	license2 := license2Value.(Map).HashMap()
	assertFields := func(name string, fields map[string]Value, want ...string) {
		t.Helper()
		if len(fields) != len(want) {
			t.Fatalf("%s projection has %d fields (%v), want %d (%v)", name, len(fields), fields, len(want), want)
		}
		for _, field := range want {
			if fields[field] == nil {
				t.Fatalf("%s projection omitted %q: %v", name, field, fields)
			}
		}
	}

	assertFields("license1", license1, "domain1", "product", "owner")
	assertFields("license2", license2, "domain2", "product", "plan")
	if bytes.Equal(license1Bytes, license2Bytes) {
		t.Fatal("independent selectors produced identical projections")
	}
}

func TestNonValueCodecsRejectUnsupportedTagOptions(t *testing.T) {
	type jsonString struct {
		Count int `json:"count,string"`
	}
	type msgpackIntern struct {
		Name string `msgpack:"name,intern"`
	}

	if _, err := JSONValueCodec().Marshal(jsonString{Count: 1}); err == nil {
		t.Fatal("JSON codec silently accepted unsupported string option")
	}
	if _, err := MsgpackValueCodec().Marshal(msgpackIntern{Name: "ann"}); err == nil {
		t.Fatal("MessagePack codec silently accepted unsupported intern option")
	}
}

func TestLegacyValueSigningOptionsStayDefaultOnly(t *testing.T) {
	type legacy struct {
		Name string `value:"name,legacy" json:"name" msgpack:"name"`
	}
	record := legacy{Name: "ann"}
	if _, err := DefaultValueCodec().SignBytes(record, "legacy"); err != nil {
		t.Fatal(err)
	}
	for _, codec := range []ValueCodec{JSONValueCodec(), MsgpackValueCodec()} {
		projection, err := codec.SignBytes(record, "legacy")
		if err != nil {
			t.Fatal(err)
		}
		v, err := Unpack(projection, true)
		if err != nil {
			t.Fatal(err)
		}
		if v.(Map).Len() != 0 {
			t.Fatal("legacy value selector leaked into another tag profile")
		}
	}
}
