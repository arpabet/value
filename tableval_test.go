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

package val_test

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/shvid/val"
	"reflect"
	"encoding/json"
)

func TestEmptyTable(t *testing.T) {

	b := val.List()

	require.Equal(t, val.TABLE, b.Kind())
	require.Equal(t, val.LIST, b.Type())
	require.Equal(t, "val.tableValue", b.Class().String())
	require.Equal(t, 0, b.Len())
	require.Equal(t, 0, b.Size())
	require.Equal(t, "90", val.Hex(b))
	require.Equal(t, "[]", val.Json(b))
	require.Equal(t, "[]", b.String())

	b = val.Map()

	require.Equal(t, val.TABLE, b.Kind())
	require.Equal(t, val.MAP, b.Type())
	require.Equal(t, "val.tableValue", b.Class().String())
	require.Equal(t, 0, b.Len())
	require.Equal(t, 0, b.Size())
	require.Equal(t, "80", val.Hex(b))
	require.Equal(t, "{}", val.Json(b))
	require.Equal(t, "{}", b.String())

}

func TestConvertTable(t *testing.T) {

	b := val.List()
	require.True(t,  b.Sorted())

	require.Equal(t, val.TABLE, b.Kind())
	require.Equal(t, val.LIST, b.Type())
	require.Equal(t, 0, b.Len())
	require.Equal(t, 0, b.Size())

	b.Put("first", val.Long(1))
	require.False(t,  b.Sorted())
	require.Equal(t, 1, b.Len())

	require.Equal(t, val.TABLE, b.Kind())
	require.Equal(t, val.MAP, b.Type())

	require.Equal(t, 1, b.Size())

	// Clear

	b.Clear()
	require.True(t,  b.Sorted())

	require.Equal(t, val.TABLE, b.Kind())
	require.Equal(t, val.LIST, b.Type())

	require.Equal(t, 0, b.Len())
	require.Equal(t, 0, b.Size())

}

func TestTableInsert(t *testing.T) {

	b := val.List()
	require.Equal(t, 0, b.MaxIndex())

	b.Insert(val.Long(123))
	require.False(t,  b.Sorted())

	require.Equal(t, val.LIST, b.Type())
	require.Equal(t, 1, b.Len())

	require.Equal(t, 1, b.MaxIndex())
}

func TestTablePutAt(t *testing.T) {

	b := val.List()
	require.Equal(t, 0, b.MaxIndex())

	b.PutAt(7, val.Long(777))
	b.PutAt(9, val.Long(999))
	b.PutAt(5, val.Long(555))
	require.False(t,  b.Sorted())
	require.Equal(t,  9, b.MaxIndex())

	// Get

	require.True(t, val.Long(555).Equal(b.GetAt(5)))
	require.True(t, val.Long(777).Equal(b.GetAt(7)))
	require.True(t, val.Long(999).Equal(b.GetAt(9)))
	require.False(t, val.Long(0).Equal(b.GetAt(0)))

	// Replace

	b.PutAt(9, val.Utf8("*"))
	require.True(t, val.Utf8("*").Equal(b.GetAt(9)))

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
	expectedMap := map[string]val.Value {
		"5": val.Long(555),
	}
	require.True(t, reflect.DeepEqual(expectedMap, b.Map()))

	// Test List
	expectedList := []val.Value {
		val.Long(555),
	}
	require.True(t, reflect.DeepEqual(expectedList, b.List()))

}

func TestTablePut(t *testing.T) {

	b := val.Map()

	b.Put("name", val.Utf8("alex"))
	b.Put("state", val.Utf8("CA"))
	b.Put("age", val.Long(38))
	b.Put("33", val.Long(33))
	require.False(t,  b.Sorted())

	require.Equal(t, 4, b.Len())
	require.Equal(t, 4, b.Size())

	// Get

	require.True(t, val.Utf8("alex").Equal(b.Get("name")))
	require.True(t, val.Utf8("CA").Equal(b.Get("state")))
	require.True(t, val.Long(38).Equal(b.Get("age")))
	require.True(t, val.Long(33).Equal(b.Get("33")))
	require.True(t, val.Long(33).Equal(b.GetAt(33)))

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
	expectedMap := map[string]val.Value {
		"33": val.Long(33),
		"name": val.Utf8("alex"),
	}
	require.True(t, reflect.DeepEqual(expectedMap, b.Map()))

	// Test List
	expectedList := []val.Value {
		val.Long(33),
		val.Utf8("alex"),
	}
	require.True(t, reflect.DeepEqual(expectedList, b.List()))

}

func TestTablePutLongNum(t *testing.T) {

	b := val.List()

	b.Put("12345678901234567890", val.Long(555))

	require.Equal(t, val.MAP, b.Type())

	num := b.GetNumber("12345678901234567890")
	require.NotNil(t, num)

	require.True(t, val.Long(555).Equal(num))

	b.Remove("12345678901234567890")

	require.Equal(t, 0, b.Size())

}

func TestJsonListTable(t *testing.T) {

	b := val.List()

	b.Insert(val.Boolean(true))
	b.Insert(val.Long(123))
	b.Insert(val.Double(-12.34))
	b.Insert(val.Utf8("text"))
	b.Insert(val.Raw([]byte{0, 1, 2}, false))

	require.Equal(t, "[true,123,-12.34,\"text\",\"base64,AAEC\"]", val.Json(b))
	require.Equal(t, "95c37bcbc028ae147ae147aea474657874c403000102", val.Hex(b))

	testPackUnpack(t, b)
}

func TestJsonMapTable(t *testing.T) {

	b := val.Map()

	b.Insert(val.Boolean(true))
	b.Insert(val.Long(123))
	b.Insert(val.Double(-12.34))
	b.Insert(val.Utf8("text"))
	b.Insert(val.Raw([]byte{0, 1, 2}, false))

	require.Equal(t, "{\"1\": true,\"2\": 123,\"3\": -12.34,\"4\": \"text\",\"5\": \"base64,AAEC\"}", val.Json(b))
	require.Equal(t, "8501c3027b03cbc028ae147ae147ae04a47465787405c403000102", val.Hex(b))

	b = val.Map()

	c := val.Map()
	c.Put("5", val.Long(5))

	b.Put("name", val.Utf8("name"))
	b.Put("123", val.Long(123))
	b.Put("map", c)

	require.Equal(t,  "{\"123\": 123,\"map\": {\"5\": 5},\"name\": \"name\"}", val.Json(b))
	require.Equal(t, "837b7ba36d6170810505a46e616d65a46e616d65", val.Hex(b))

	testPackUnpack(t, b)
}

func TestCycleTable(t *testing.T) {

	b := val.Map()
	b.Put("map", b)

	require.Equal(t,  "{\"map\": null}", val.Json(b))
	require.Equal(t, "81a36d6170c0", val.Hex(b))
}

type testTableStruct struct {
	T val.Table
}

func TestTableMarshal(t *testing.T) {

	b := val.List()
	b.Insert(val.Long(100))

	j, _ := b.MarshalJSON()
	require.Equal(t, "[100]", string(j))

	bin, _ := b.MarshalBinary()
	require.Equal(t, []byte{0x91, 0x64}, bin)

	b = val.Map()
	b.Put("a", val.Boolean(true))

	j, _ = b.MarshalJSON()
	require.Equal(t, "{\"a\": true}", string(j))

	bin, _ = b.MarshalBinary()
	require.Equal(t,  []byte{0x81, 0xa1, 0x61, 0xc3}, bin)

	s := &testTableStruct{b}

	j, _ = json.Marshal(s)
	require.Equal(t, "{\"T\":{\"a\":true}}", string(j))

}