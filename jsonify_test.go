/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value_test

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	val "go.arpabet.com/value"
)

// esc is the escape of a UTF-16 code unit as it stands in a JSON text: a
// backslash, u and four hexadecimal digits.
func esc(unit int) string {
	return fmt.Sprintf("%cu%04x", '\\', unit)
}

// quoted is a JSON string of the parts.
func quoted(parts ...string) string {
	out := `"`
	for _, p := range parts {
		out += p
	}
	return out + `"`
}

const (
	replacement = 0xfffd  // U+FFFD, what a byte that is not UTF-8 reads back as
	zeroWidth   = 0x200b  // U+200B, in the basic plane and not printable
	languageTag = 0xe0001 // U+E0001, beyond the basic plane and not printable
)

// The texts Jsonify gave no JSON for before v1.6.2, and what it writes now.
var jsonifyRepaired = []struct{ text, json string }{
	{"\a", quoted(esc(0x07))},
	{"\v", quoted(esc(0x0b))},
	{"\x00", quoted(esc(0x00))},
	{"\x1b[0m", quoted(esc(0x1b), "[0m")},
	{"\x7f", quoted(esc(0x7f))},
	{string(rune(languageTag)), quoted(esc(0xdb40), esc(0xdc01))},                  // its surrogate pair
	{"\xff", quoted(esc(replacement))},                                             // a byte that is not UTF-8
	{"a\xc3", quoted("a", esc(replacement))},                                       // a sequence cut short
	{"\xed\xa0\x80", quoted(esc(replacement), esc(replacement), esc(replacement))}, // a surrogate, which UTF-8 does not encode
}

// Texts whose rendering was JSON already: it is what it was, strconv.Quote.
var jsonifyKept = []string{
	"",
	"text",
	`quote " and backslash \`,
	"line\nreturn\rtab\tbackspace\bfeed\f",
	"h" + string(rune(0xe9)) + "llo",       // a Latin letter with an accent
	string([]rune{0x65e5, 0x672c, 0x8a9e}), // CJK
	string(rune(0x1f600)),                  // printable, beyond the basic plane
	"zero" + string(rune(zeroWidth)) + "width", // escaped, as it was
	string(rune(replacement)),                  // the replacement character itself is printable
	"{\"a\": [1,2]}",
}

// decoded is the text a JSON reader gets back: s, with U+FFFD for every byte
// that is not UTF-8.
func decoded(s string) string {
	return string([]rune(s))
}

func TestJsonifyStringIsJSON(t *testing.T) {
	for _, c := range jsonifyRepaired {
		s := val.Utf8(c.text)
		require.Equal(t, c.json, val.Jsonify(s), "%q", c.text)
		require.Equal(t, c.text, s.String(), "String() is the text itself")

		var back string
		require.NoError(t, json.Unmarshal([]byte(val.Jsonify(s)), &back), "%q", c.text)
		require.Equal(t, decoded(c.text), back)

		// MarshalJSON gave encoding/json a text it refused.
		out, err := json.Marshal(s)
		require.NoError(t, err, "%q", c.text)
		require.NoError(t, json.Unmarshal(out, &back))
		require.Equal(t, decoded(c.text), back)
	}
}

func TestJsonifyStringKeepsItsRendering(t *testing.T) {
	for _, text := range jsonifyKept {
		require.True(t, json.Valid([]byte(strconv.Quote(text))), "%q", text)
		require.Equal(t, strconv.Quote(text), val.Jsonify(val.Utf8(text)), "%q", text)

		var back string
		require.NoError(t, json.Unmarshal([]byte(val.Jsonify(val.Utf8(text))), &back))
		require.Equal(t, text, back)
	}
	// A character that is not printable stays escaped, in the basic plane as before.
	require.Equal(t, quoted("zero", esc(zeroWidth), "width"), val.Jsonify(val.Utf8("zero"+string(rune(zeroWidth))+"width")))
}

// A key was written between quotes as it was, so a quote in it ended it.
func TestJsonifyMapKeyIsJSON(t *testing.T) {
	keys := []string{`a"b`, `back\slash`, "line\nbreak", "\x00", "\xff", "plain"}
	entries := make(map[string]val.Value, len(keys))
	want := make(map[string]int64, len(keys))
	for i, k := range keys {
		entries[k] = val.Long(int64(i))
		want[decoded(k)] = int64(i)
	}
	for name, m := range map[string]val.Map{
		"immutable": val.ImmutableMapOf(entries),
		"sorted":    val.SortedMapOf(entries),
	} {
		var back map[string]int64
		require.NoError(t, json.Unmarshal([]byte(val.Jsonify(m)), &back), name)
		require.Equal(t, want, back, name)
		require.Equal(t, val.Jsonify(m), m.String(), name)

		out, err := json.Marshal(m)
		require.NoError(t, err, name)
		back = nil
		require.NoError(t, json.Unmarshal(out, &back), name)
		require.Equal(t, want, back, name)
	}

	// A key that needs no escape is written as before.
	key := string([]rune{0x43a, 0x43b, 0x44e, 0x447}) // Cyrillic
	m := val.ImmutableMapOf(map[string]val.Value{"name": val.Utf8("ann"), key: val.Long(1)})
	require.Equal(t, `{"name": "ann","`+key+`": 1}`, val.Jsonify(m))
}

func TestJsonifyDoubleThatIsNotFinite(t *testing.T) {
	for _, d := range []float64{math.Inf(1), math.Inf(-1), math.NaN()} {
		n := val.Double(d)
		require.Equal(t, "null", val.Jsonify(n))
		out, err := json.Marshal(n)
		require.NoError(t, err)
		require.Equal(t, "null", string(out))
	}
	// String() is not JSON and says what the double is.
	require.Equal(t, "+Inf", val.Double(math.Inf(1)).String())
	require.Equal(t, "-Inf", val.Double(math.Inf(-1)).String())

	l := val.ImmutableList([]val.Value{val.Double(math.Inf(1)), val.Double(1.5), val.Double(math.Inf(-1))})
	require.Equal(t, "[null,1.5,null]", val.Jsonify(l))
	require.True(t, json.Valid([]byte(val.Jsonify(l))))
}

// FuzzJsonify: whatever the text, as a value and as a key, Jsonify writes JSON
// that reads back as the text, and writes what strconv.Quote writes wherever
// that is JSON.
func FuzzJsonify(f *testing.F) {
	for _, c := range jsonifyRepaired {
		f.Add(c.text)
	}
	for _, text := range jsonifyKept {
		f.Add(text)
	}
	f.Fuzz(func(t *testing.T, text string) {
		got := val.Jsonify(val.Utf8(text))
		var back string
		if err := json.Unmarshal([]byte(got), &back); err != nil {
			t.Fatalf("Jsonify(%q) = %s is not JSON: %v", text, got, err)
		}
		if back != decoded(text) {
			t.Fatalf("Jsonify(%q) = %s reads back as %q", text, got, back)
		}
		if was := strconv.Quote(text); json.Valid([]byte(was)) && got != was {
			t.Fatalf("Jsonify(%q) = %s, was %s", text, got, was)
		}

		m := val.ImmutableMapOf(map[string]val.Value{text: val.Null})
		var keys map[string]interface{}
		if err := json.Unmarshal([]byte(val.Jsonify(m)), &keys); err != nil {
			t.Fatalf("Jsonify of a map with the key %q = %s is not JSON: %v", text, val.Jsonify(m), err)
		}
		if _, ok := keys[decoded(text)]; !ok || len(keys) != 1 {
			t.Fatalf("Jsonify of a map with the key %q = %s reads back as %v", text, val.Jsonify(m), keys)
		}
	})
}
