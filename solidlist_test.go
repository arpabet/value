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
	val "github.com/consensusdb/value"
	"github.com/stretchr/testify/require"
	"testing"
)

/**
	@author Alex Shvid
*/

func TestNilSolidList(t *testing.T) {

	b := val.EmptyList()
	b = b.Append(nil)

	data, err := val.Pack(b)
	require.Nil(t, err)

	actual, err := val.Unpack(data, false)
	require.Nil(t, err)
	require.Equal(t, val.LIST, actual.Kind())

	tbl := actual.(val.List)
	require.Equal(t, 1, tbl.Len())

	testPackUnpack(t, b)
}

func TestEmptySolidList(t *testing.T) {

	b := val.EmptyList()

	require.Equal(t, val.LIST, b.Kind())
	require.Equal(t, "value.solidListValue", b.Class().String())
	require.Equal(t, 0, b.Len())
	require.Equal(t, "90", val.Hex(b))
	require.Equal(t, "[]", val.Jsonify(b))
	require.Equal(t, "[]", b.String())

}

func TestSolidListAppend(t *testing.T) {

	b := val.EmptyList()
	b = b.Append(val.Long(123))

	require.Equal(t, val.LIST, b.Kind())
	require.Equal(t, 1, b.Len())

}


func TestSolidListPutAt(t *testing.T) {

	b := val.EmptyList()

	b = b.PutAt(7, val.Long(777))
	b = b.PutAt(9, val.Long(999))
	b = b.PutAt(5, val.Long(555))
	require.Equal(t,  10, b.Len())

	// Get

	require.True(t, val.Long(555).Equal(b.GetAt(5)))
	require.True(t, val.Long(777).Equal(b.GetAt(7)))
	require.True(t, val.Long(999).Equal(b.GetAt(9)))
	require.Nil(t, b.GetAt(0))
	require.Nil(t, b.GetAt(1))
	require.Nil(t, b.GetAt(2))

	// Replace

	b = b.PutAt(9, val.Utf8("*"))
	require.True(t, val.Utf8("*").Equal(b.GetAt(9)))

	// Remove

	b = b.RemoveAt(7)
	require.Equal(t, 9, b.Len())

	b = b.RemoveAt(8)
	require.Equal(t, 8, b.Len())

}

func TestSolidListMarshal(t *testing.T) {

	b := val.EmptyList()
	b = b.Append(val.Long(100))

	j, _ := b.MarshalJSON()
	require.Equal(t, "[100]", string(j))

	bin, _ := b.MarshalBinary()
	require.Equal(t, []byte{0x91, 0x64}, bin)

	b = val.EmptyList()
	b = b.PutAt(3, val.Boolean(true))

	j, _ = b.MarshalJSON()
	require.Equal(t, "[null,null,null,true]", string(j))

	bin, _ = b.MarshalBinary()
	require.Equal(t,  []byte{0x94, 0xc0, 0xc0, 0xc0, 0xc3}, bin)


}

func TestSolidListJson(t *testing.T) {

	b := val.EmptyList()

	b = b.Append(val.Boolean(true))
	b = b.Append(val.Long(123))
	b = b.Append(val.Double(-12.34))
	b = b.Append(val.Utf8("text"))
	b = b.Append(val.Raw([]byte{0, 1, 2}, false))

	require.Equal(t, "[true,123,-12.34,\"text\",\"base64,AAEC\"]", val.Jsonify(b))
	require.Equal(t, "95c37bcbc028ae147ae147aea474657874c403000102", val.Hex(b))

	testPackUnpack(t, b)

}
