/*
 *
 * Copyright 2020-present Arpabet, Inc.
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

package value_test

import (
	val "arpabet.pkg.is/value"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

/**
	@author Alex Shvid
*/


func TestNilSortedMap(t *testing.T) {

	b := val.EmptyMap()
	b = b.Put("null", nil)

	data, err := val.Pack(b)
	require.Nil(t, err)

	actual, err := val.Unpack(data, false)
	require.Nil(t, err)
	require.Equal(t, val.MAP, actual.Kind())

	tbl := actual.(val.Map)
	require.Equal(t, 1, tbl.Len())

	testPackUnpack(t, b)
}

func TestEmptySortedMap(t *testing.T) {

	b := val.EmptyMap()

	require.Equal(t, val.MAP, b.Kind())
	require.Equal(t, "value.sortedMapValue", b.Class().String())
	require.Equal(t, 0, b.Len())
	require.Equal(t, "80", val.Hex(b))
	require.Equal(t, "{}", val.Jsonify(b))
	require.Equal(t, "{}", b.String())

}

func TestSortedMapPut(t *testing.T) {

	b := val.EmptyMap()

	b = b.Put("name", val.Utf8("alex"))
	b = b.Put("state", val.Utf8("CA"))
	b = b.Put("age", val.Long(38))
	b = b.Put("33", val.Long(33))

	require.Equal(t, 4, b.Len())


	// Get

	require.True(t, val.Utf8("alex").Equal(b.GetString("name")))
	require.True(t, val.Utf8("CA").Equal(b.GetString("state")))
	require.True(t, val.Long(38).Equal(b.GetNumber("age")))
	require.True(t, val.Long(33).Equal(b.GetNumber("33")))

	// Insert
	b = b.Insert("33", val.Long(33))
	require.Equal(t, 5, b.Len())
	require.True(t, val.Long(33).Equal(b.GetNumber("33")))

	// Remove 33
	b = b.Remove("33")
	require.Equal(t, 4, b.Len())
	require.True(t, val.Long(33).Equal(b.GetNumber("33")))

	// Remove
	b = b.Remove("age")
	require.Equal(t, 3, b.Len())

	require.Equal(t, []string{"33", "name", "state"}, b.Keys())

	// Remove
	b = b.Remove("state")
	require.Equal(t, 2, b.Len())

	// Test Map
	expectedMap := map[string]val.Value {
		"33": val.Long(33),
		"name": val.Utf8("alex"),
	}
	require.True(t, reflect.DeepEqual(expectedMap, b.HashMap()))

	// Test List
	expectedList := []val.Value {
		val.Long(33),
		val.Utf8("alex"),
	}
	require.True(t, reflect.DeepEqual(expectedList, b.Values()))

}

func TestSortedMapMarshal(t *testing.T) {

	b := val.EmptyMap()
	b = b.Put("k", val.Long(100))

	j, _ := b.MarshalJSON()
	require.Equal(t, "{\"k\": 100}", string(j))

	bin, _ := b.MarshalBinary()
	require.Equal(t, []byte{0x81, 0xa1, 0x6b, 0x64}, bin)

	b = val.EmptyMap()
	b = b.Put("3", val.Boolean(true))

	j, _ = b.MarshalJSON()
	require.Equal(t, "{\"3\": true}", string(j))

	bin, _ = b.MarshalBinary()
	require.Equal(t,  []byte{0x81, 0xa1, 0x33, 0xc3}, bin)


}

func TestSortedMapPutLongNum(t *testing.T) {

	b := val.EmptyMap()

	b = b.Put("12345678901234567890", val.Long(555))

	require.Equal(t, val.MAP, b.Kind())

	num := b.GetNumber("12345678901234567890")
	require.NotNil(t, num)

	require.True(t, val.Long(555).Equal(num))

	b = b.Remove("12345678901234567890")

	require.Equal(t, 0, b.Len())

}

func TestSortedMapJson(t *testing.T) {

	d := val.EmptyMap()

	c := val.EmptyMap()
	c = c.Put("5", val.Long(5))

	d = d.Put("name", val.Utf8("name"))
	d = d.Put("123", val.Long(123))
	d = d.Put("map", c)

	require.Equal(t,  "{\"123\": 123,\"map\": {\"5\": 5},\"name\": \"name\"}", val.Jsonify(d))
	require.Equal(t, "83a33132337ba36d617081a13505a46e616d65a46e616d65", val.Hex(d))

	testPackUnpack(t, d)

}


func newHandshakeRequest(clientId int64) val.Map {
	return val.EmptyMap().
		Put("m", val.Utf8("vRPC")).
		Put("v", val.Double(1.0)).
		Put("t", val.Long(1)).
		Put("rid", val.Long(123)).
		Put("cid", val.Long(clientId))
}

func TestSortedMapProtocol(t *testing.T) {

	req := newHandshakeRequest(555)
	require.Equal(t,  5, req.Len())
	require.Equal(t,  "{\"cid\": 555,\"m\": \"vRPC\",\"rid\": 123,\"t\": 1,\"v\": 1}", req.String())

}