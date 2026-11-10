/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value_test

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"os"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	val "go.arpabet.com/value"
)

// vectorValue is the language-neutral notation of testdata/canonical.json.
type vectorValue struct {
	Kind    string                 `json:"kind"`
	Value   bool                   `json:"value"`
	Text    string                 `json:"text"`
	Hex     string                 `json:"hex"`
	Seconds int64                  `json:"seconds"`
	Nanos   int64                  `json:"nanos"`
	Tag     int                    `json:"tag"`
	Items   []vectorValue          `json:"items"`
	Entries map[string]vectorValue `json:"entries"`
}

type vector struct {
	Name  string      `json:"name"`
	Value vectorValue `json:"value"`
	Hex   string      `json:"hex"`
	Text  string      `json:"text"`
	JSON  string      `json:"json"`
	// JSONCanonical is the lossless JSON in canonical mode.
	JSONCanonical string `json:"jsonCanonical"`
}

func buildVector(t *testing.T, v vectorValue) val.Value {
	switch v.Kind {
	case "null":
		return val.Null
	case "bool":
		return val.Boolean(v.Value)
	case "long":
		n, ok := new(big.Int).SetString(v.Text, 10)
		require.True(t, ok, v.Text)
		require.True(t, n.IsInt64(), v.Text)
		return val.Long(n.Int64())
	case "double":
		d, err := decimal.NewFromString(v.Text)
		require.NoError(t, err)
		f, _ := d.Float64()
		return val.Double(f)
	case "bigint":
		n, ok := new(big.Int).SetString(v.Text, 10)
		require.True(t, ok, v.Text)
		return val.BigInt(n)
	case "decimal":
		d, err := decimal.NewFromString(v.Text)
		require.NoError(t, err)
		return val.Decimal(d)
	case "str":
		return val.Utf8(v.Text)
	case "bin":
		b, err := hex.DecodeString(v.Hex)
		require.NoError(t, err)
		return val.Raw(b, false)
	case "timestamp":
		return val.UnixTime(v.Seconds, v.Nanos)
	case "ext":
		b, err := hex.DecodeString(v.Hex)
		require.NoError(t, err)
		return val.Unknown(append([]byte{byte(v.Tag)}, b...))
	case "list":
		items := make([]val.Value, len(v.Items))
		for i, item := range v.Items {
			items[i] = buildVector(t, item)
		}
		return val.ImmutableList(items)
	case "map":
		m := make(map[string]val.Value, len(v.Entries))
		for k, e := range v.Entries {
			m[k] = buildVector(t, e)
		}
		return val.ImmutableMapOf(m)
	default:
		t.Fatalf("unknown vector kind %q", v.Kind)
		return nil
	}
}

// TestCanonicalVectors checks testdata/canonical.json, the vectors every
// implementation of the canonical form must reproduce: the bytes of each
// value, that the bytes decode to an equal value, that decoding is a fixed
// point, the String and JSON renderings where the file specifies them, and
// the canonical lossless JSON of every value.
func TestCanonicalVectors(t *testing.T) {
	data, err := os.ReadFile("testdata/canonical.json")
	require.NoError(t, err)
	var file struct {
		Vectors []vector `json:"vectors"`
	}
	require.NoError(t, json.Unmarshal(data, &file))
	require.NotEmpty(t, file.Vectors)

	seen := map[string]bool{}
	for _, vec := range file.Vectors {
		require.False(t, seen[vec.Name], "duplicate vector %q", vec.Name)
		seen[vec.Name] = true
		v := buildVector(t, vec.Value)
		require.Equal(t, vec.Hex, val.Hex(v), "encoding of %q", vec.Name)

		b, err := hex.DecodeString(vec.Hex)
		require.NoError(t, err)
		back, err := val.Unpack(b, false)
		require.NoError(t, err, vec.Name)
		require.True(t, v.Equal(back), "%q decodes to %s, want %s", vec.Name, back, v)
		require.Equal(t, vec.Hex, val.Hex(back), "re-encoding of %q", vec.Name)

		if vec.Text != "" {
			require.Equal(t, vec.Text, v.String(), "String of %q", vec.Name)
		}
		if vec.JSON != "" {
			require.Equal(t, vec.JSON, val.Jsonify(v), "JSON of %q", vec.Name)
		}
		require.NotEmpty(t, vec.JSONCanonical, "%q has no jsonCanonical", vec.Name)
		text, err := val.FormatJSON(v, val.JSONCanonical)
		require.NoError(t, err, vec.Name)
		require.Equal(t, vec.JSONCanonical, string(text), "canonical JSON of %q", vec.Name)
		parsed, err := val.ParseJSON([]byte(vec.JSONCanonical))
		require.NoError(t, err, vec.Name)
		require.Equal(t, vec.Hex, val.Hex(parsed), "canonical JSON of %q read back", vec.Name)
	}
}
