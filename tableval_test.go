package genval_test

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/shvid/genval"
)

func TestEmptyTable(t *testing.T) {

	b := genval.List()

	require.Equal(t, genval.TABLE, b.Kind())
	require.Equal(t, genval.LIST, b.Type())
	require.Equal(t, "genval.tableValue", b.Class().String())
	require.Equal(t, 0, b.Len())
	require.Equal(t, 0, b.Size())
	require.Equal(t, "c0", genval.Hex(b))
	require.Equal(t, "{}", b.Json())
	require.Equal(t, "{}", b.String())

	b = genval.Map()

	require.Equal(t, genval.TABLE, b.Kind())
	require.Equal(t, genval.MAP, b.Type())
	require.Equal(t, "genval.tableValue", b.Class().String())
	require.Equal(t, 0, b.Len())
	require.Equal(t, 0, b.Size())
	require.Equal(t, "c0", genval.Hex(b))
	require.Equal(t, "{}", b.Json())
	require.Equal(t, "{}", b.String())

}

func TestConvertTable(t *testing.T) {

	b := genval.List()
	require.True(t,  b.Sorted())

	require.Equal(t, genval.TABLE, b.Kind())
	require.Equal(t, genval.LIST, b.Type())
	require.Equal(t, 0, b.Len())
	require.Equal(t, 0, b.Size())

	b.Put("first", genval.Long(1))
	require.False(t,  b.Sorted())

	require.Equal(t, genval.TABLE, b.Kind())
	require.Equal(t, genval.MAP, b.Type())

	require.Equal(t, 1, b.Len())
	require.Equal(t, 1, b.Size())

	// Clear

	b.Clear()
	require.True(t,  b.Sorted())

	require.Equal(t, genval.TABLE, b.Kind())
	require.Equal(t, genval.LIST, b.Type())

	require.Equal(t, 0, b.Len())
	require.Equal(t, 0, b.Size())

}

func TestTableInsert(t *testing.T) {

	b := genval.List()
	require.Equal(t, 0, b.MaxIndex())

	b.Insert(genval.Long(123))
	require.False(t,  b.Sorted())

	require.Equal(t, genval.LIST, b.Type())
	require.Equal(t, 1, b.Len())

	require.Equal(t, 1, b.MaxIndex())
}

func TestTablePutAt(t *testing.T) {

	b := genval.List()
	require.Equal(t, 0, b.MaxIndex())

	b.PutAt(7, genval.Long(777))
	b.PutAt(9, genval.Long(999))
	b.PutAt(5, genval.Long(555))
	require.False(t,  b.Sorted())
	require.Equal(t,  9, b.MaxIndex())

	require.True(t, genval.Long(555).Equal(b.GetAt(5)))
	require.True(t, genval.Long(777).Equal(b.GetAt(7)))
	require.True(t, genval.Long(999).Equal(b.GetAt(9)))
	require.False(t, genval.Long(0).Equal(b.GetAt(0)))

	require.Equal(t, []int{5, 7, 9}, b.Indexes())
	require.True(t,  b.Sorted())
	require.Equal(t, []string{"5", "7", "9"}, b.Keys())
	require.Equal(t, 3, b.Size())

	b.RemoveAt(7)
	require.False(t,  b.Sorted())
	require.Equal(t, 4, b.Len())
	require.Equal(t, 2, b.Size())
	require.Equal(t, []int{5, 9}, b.Indexes())
	require.Equal(t, []string{"5", "9"}, b.Keys())

	b.RemoveAt(9)
	require.Equal(t,  9, b.MaxIndex())

}