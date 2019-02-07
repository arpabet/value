package genval_test

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/shvid/genval"
	"reflect"
)

func TestEmptyTable(t *testing.T) {

	b := genval.List()

	require.Equal(t, genval.TABLE, b.Kind())
	require.Equal(t, genval.LIST, b.Type())
	require.Equal(t, "genval.tableValue", b.Class().String())
	require.Equal(t, 0, b.Len())
	require.Equal(t, 0, b.Size())
	require.Equal(t, "90", genval.Hex(b))
	require.Equal(t, "[]", b.Json())
	require.Equal(t, "[]", b.String())

	b = genval.Map()

	require.Equal(t, genval.TABLE, b.Kind())
	require.Equal(t, genval.MAP, b.Type())
	require.Equal(t, "genval.tableValue", b.Class().String())
	require.Equal(t, 0, b.Len())
	require.Equal(t, 0, b.Size())
	require.Equal(t, "80", genval.Hex(b))
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
	require.Equal(t, 1, b.Len())

	require.Equal(t, genval.TABLE, b.Kind())
	require.Equal(t, genval.MAP, b.Type())

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

	// Get

	require.True(t, genval.Long(555).Equal(b.GetAt(5)))
	require.True(t, genval.Long(777).Equal(b.GetAt(7)))
	require.True(t, genval.Long(999).Equal(b.GetAt(9)))
	require.False(t, genval.Long(0).Equal(b.GetAt(0)))

	// Replace

	b.PutAt(9, genval.Utf8("*"))
	require.True(t, genval.Utf8("*").Equal(b.GetAt(9)))

	require.Equal(t, []int{5, 7, 9}, b.Indexes())
	require.True(t,  b.Sorted())
	require.Equal(t, []string{"5", "7", "9"}, b.Keys())
	require.Equal(t, 3, b.Size())

	// Remove

	b.RemoveAt(7)
	require.False(t,  b.Sorted())
	require.Equal(t, 5, b.Len())
	require.Equal(t, 2, b.Size())
	require.Equal(t, []int{5, 9}, b.Indexes())
	require.Equal(t, []string{"5", "9"}, b.Keys())

	b.RemoveAt(9)
	require.Equal(t,  9, b.MaxIndex())
	require.Equal(t, 6, b.Len())
	require.Equal(t, 1, b.Size())

	// Test Map
	expectedMap := map[string]genval.Value {
		"5": genval.Long(555),
	}
	require.True(t, reflect.DeepEqual(expectedMap, b.Map()))

	// Test List
	expectedList := []genval.Value {
		genval.Long(555),
	}
	require.True(t, reflect.DeepEqual(expectedList, b.List()))

}

func TestTablePut(t *testing.T) {

	b := genval.Map()

	b.Put("name", genval.Utf8("alex"))
	b.Put("state", genval.Utf8("CA"))
	b.Put("age", genval.Long(38))
	b.Put("33", genval.Long(33))
	require.False(t,  b.Sorted())

	require.Equal(t, 4, b.Len())
	require.Equal(t, 4, b.Size())

	// Get

	require.True(t, genval.Utf8("alex").Equal(b.Get("name")))
	require.True(t, genval.Utf8("CA").Equal(b.Get("state")))
	require.True(t, genval.Long(38).Equal(b.Get("age")))
	require.True(t, genval.Long(33).Equal(b.Get("33")))
	require.True(t, genval.Long(33).Equal(b.GetAt(33)))

	// Remove

	b.Remove("age")
	require.Equal(t, 5, b.Len())
	require.Equal(t, 3, b.Size())

	require.Equal(t, []int{33}, b.Indexes())
	require.Equal(t, []string{"33", "name", "state"}, b.Keys())
	require.Equal(t,  33, b.MaxIndex())

	// Remove
	b.Remove("state")
	require.Equal(t, 6, b.Len())
	require.Equal(t, 2, b.Size())

	// Test Map
	expectedMap := map[string]genval.Value {
		"33": genval.Long(33),
		"name": genval.Utf8("alex"),
	}
	require.True(t, reflect.DeepEqual(expectedMap, b.Map()))

	// Test List
	expectedList := []genval.Value {
		genval.Long(33),
		genval.Utf8("alex"),
	}
	require.True(t, reflect.DeepEqual(expectedList, b.List()))

}

func TestTablePutLongNum(t *testing.T) {

	b := genval.List()

	b.Put("12345678901234567890", genval.Long(555))

	require.Equal(t, genval.MAP, b.Type())

	num := b.GetNumber("12345678901234567890")
	require.NotNil(t, num)

	require.True(t, genval.Long(555).Equal(num))

	b.Remove("12345678901234567890")

	require.Equal(t, 0, b.Size())

}