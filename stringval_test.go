package genval_test

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/shvid/genval"
)

var testStrings = map[string]string {

	"": "a0",
	"test": "a474657374",
	"hello": "a568656c6c6f",
}

func TestString(t *testing.T) {

	for str, hex := range testStrings {

		b := genval.Utf8(str)

		require.Equal(t, genval.STRING, b.Kind())
		require.Equal(t, "genval.StringValue", b.Class().String())
		require.Equal(t, hex, genval.Hex(b))
		require.Equal(t, "\""+ str + "\"", b.Json())
		require.Equal(t, str, b.String())

	}

}