/*
 *
 * Copyright 2019-present Alexander Shvid
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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

func TestJsonListTable(t *testing.T) {

	b := genval.List()

	b.Insert(genval.Boolean(true))
	b.Insert(genval.Long(123))
	b.Insert(genval.Double(-12.34))
	b.Insert(genval.Utf8("text"))
	b.Insert(genval.Raw([]byte{0, 1, 2}, false))

	require.Equal(t, "[true,123,-12.34,\"text\",\"base64,AAEC\"]", b.Json())
	require.Equal(t, "95c37bcbc028ae147ae147aea474657874c403000102", genval.Hex(b))

}

func TestJsonMapTable(t *testing.T) {

	b := genval.Map()

	b.Insert(genval.Boolean(true))
	b.Insert(genval.Long(123))
	b.Insert(genval.Double(-12.34))
	b.Insert(genval.Utf8("text"))
	b.Insert(genval.Raw([]byte{0, 1, 2}, false))

	require.Equal(t, "{\"1\": true,\"2\": 123,\"3\": -12.34,\"4\": \"text\",\"5\": \"base64,AAEC\"}", b.Json())
	require.Equal(t, "8501c3027b03cbc028ae147ae147ae04a47465787405c403000102", genval.Hex(b))

	b = genval.Map()

	c := genval.Map()
	c.Put("5", genval.Long(5))

	b.Put("name", genval.Utf8("name"))
	b.Put("123", genval.Long(123))
	b.Put("map", c)

	require.Equal(t,  "{\"123\": 123,\"map\": {\"5\": 5},\"name\": \"name\"}", b.Json())
	require.Equal(t, "837b7ba36d6170810505a46e616d65a46e616d65", genval.Hex(b))

}

func TestCycleTable(t *testing.T) {

	b := genval.Map()
	b.Put("map", b)

	require.Equal(t,  "{\"map\": {}}", b.Json())
	require.Equal(t, "81a36d6170c0", genval.Hex(b))
}