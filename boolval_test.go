package genval_test

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/shvid/genval"
)

func TestBool(t *testing.T) {

	b := genval.Boolean(true)

	require.Equal(t, genval.BOOL, b.Kind())
	require.Equal(t, "genval.BoolValue", b.Class().String())
	require.Equal(t, "c3", genval.Hex(b))
	require.Equal(t, "true", b.Json())
	require.Equal(t, "true", b.String())

	require.Equal(t, true, genval.ParseBoolean("t").Boolean())
	require.Equal(t, true, genval.ParseBoolean("true").Boolean())
	require.Equal(t, true, genval.ParseBoolean("True").Boolean())

	b = genval.Boolean(false)
	require.Equal(t, "c2", genval.Hex(b))
	require.Equal(t, "false", b.Json())
	require.Equal(t, "false", b.String())

	require.Equal(t, false, genval.ParseBoolean("f").Boolean())
	require.Equal(t, false, genval.ParseBoolean("false").Boolean())
	require.Equal(t, false, genval.ParseBoolean("False").Boolean())
	require.Equal(t, false, genval.ParseBoolean("").Boolean())
	require.Equal(t, false, genval.ParseBoolean("any_value").Boolean())

}