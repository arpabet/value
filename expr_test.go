package genval_test

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/shvid/genval"
)

func TestExpr(t *testing.T) {

	b := genval.Expression("")

	require.Equal(t, 1, b.Size())
	require.Equal(t, "", b.GetAt(0))
	require.Equal(t, "", b.String())

	b = genval.Expression("name")

	require.Equal(t, 1, b.Size())
	require.Equal(t, "name", b.GetAt(0))
	require.Equal(t, "", b.GetAt(-1))
	require.Equal(t, "", b.GetAt(1))
	require.Equal(t, "name", b.String())

	b = genval.Expression("name.first")

	require.Equal(t, 2, b.Size())
	require.Equal(t, "name", b.GetAt(0))
	require.Equal(t, "first", b.GetAt(1))
	require.Equal(t, "", b.GetAt(-1))
	require.Equal(t, "", b.GetAt(2))
	require.Equal(t, "name.first", b.String())

}