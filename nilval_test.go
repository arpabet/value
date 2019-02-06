package genval_test

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/shvid/genval"
)

func TestNil(t *testing.T) {

	b := genval.Nil()

	require.Equal(t, genval.NIL, b.Kind())
	require.Equal(t, "genval.NilValue", b.Class().String())
	require.Equal(t, "c0", genval.Hex(b))
	require.Equal(t, "null", b.Json())
	require.Equal(t, "nil", b.String())

}