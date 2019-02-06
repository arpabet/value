package genval_test

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/shvid/genval"
)

func TestTable(t *testing.T) {

	b := genval.Map()

	require.Equal(t, genval.TABLE, b.Kind())
	require.Equal(t, "genval.tableValue", b.Class().String())
	require.Equal(t, "c0", genval.Hex(b))
	require.Equal(t, "{}", b.Json())
	require.Equal(t, "{}", b.String())

}
