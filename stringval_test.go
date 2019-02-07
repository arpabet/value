package genval_test

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/shvid/genval"
	"bytes"
	"strconv"
)

var testStrings = map[string]string {

	"": "a0",
	"test": "a474657374",
	"hello": "a568656c6c6f",
}

func TestUtf8String(t *testing.T) {

	for str, hex := range testStrings {

		b := genval.Utf8(str)

		require.Equal(t, genval.STRING, b.Kind())
		require.Equal(t, genval.UTF8, b.Type())
		require.Equal(t, "genval.stringValue", b.Class().String())
		require.Equal(t, hex, genval.Hex(b))
		require.Equal(t, "\""+ str + "\"", b.Json())
		require.Equal(t, str, b.String())

	}

}

func TestJsonString(t *testing.T) {

	src := "json\"val\"json"

	s := genval.Utf8(src)

	require.Equal(t, src, s.String())
	require.Equal(t, "\"json\\\"val\\\"json\"", s.Json())

	actual, _ := strconv.Unquote(s.Json())
	require.Equal(t, src, actual)

}

func TestRawString(t *testing.T) {

	raw := []byte { 0, 1, 2, 3, 4, 5 }
	s := genval.Raw(raw, false)

	require.Equal(t, genval.STRING, s.Kind())
	require.Equal(t, genval.RAW, s.Type())
	require.Equal(t, "base64!AAECAwQF", s.String())
	require.Equal(t, "\"base64!AAECAwQF\"", s.Json())
	require.Equal(t, "c406000102030405", genval.Hex(s))
	require.Equal(t, 0, bytes.Compare(raw, s.Raw()))

	actual := genval.ParseString(s.String())
	require.Equal(t, 0, bytes.Compare(s.Raw(), actual.Raw()))

}