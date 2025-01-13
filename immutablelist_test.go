/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */


package value_test

import (
	val "go.arpabet.com/value"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNilImmutableList(t *testing.T) {

	b := val.EmptyImmutableList()
	b = b.Append(val.Null)

	data, err := val.Pack(b)
	require.Nil(t, err)

	actual, err := val.Unpack(data, false)
	require.Nil(t, err)
	require.Equal(t, val.LIST, actual.Kind())

	tbl := actual.(val.List)
	require.Equal(t, 1, tbl.Len())

	testPackUnpack(t, b)
}

func TestEmptyImmutableList(t *testing.T) {

	b := val.EmptyImmutableList()

	require.Equal(t, val.LIST, b.Kind())
	require.Equal(t, "value.immutableListValue", b.Class().String())
	require.Equal(t, 0, b.Len())
	require.Equal(t, "90", val.Hex(b))
	require.Equal(t, "[]", val.Jsonify(b))
	require.Equal(t, "[]", b.String())

}

func TestImmutableListAppend(t *testing.T) {

	b := val.EmptyImmutableList()
	b = b.Append(val.Long(123))

	require.Equal(t, val.LIST, b.Kind())
	require.Equal(t, 1, b.Len())

}


func TestImmutableListPutAt(t *testing.T) {

	b := val.EmptyImmutableList()

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

func TestImmutableListMarshal(t *testing.T) {

	b := val.EmptyImmutableList()
	b = b.Append(val.Long(100))

	j, _ := b.MarshalJSON()
	require.Equal(t, "[100]", string(j))

	bin, _ := b.MarshalBinary()
	require.Equal(t, []byte{0x91, 0x64}, bin)

	b = val.EmptyImmutableList()
	b = b.PutAt(3, val.Boolean(true))

	j, _ = b.MarshalJSON()
	require.Equal(t, "[null,null,null,true]", string(j))

	bin, _ = b.MarshalBinary()
	require.Equal(t,  []byte{0x94, 0xc0, 0xc0, 0xc0, 0xc3}, bin)


}

func TestImmutableListJson(t *testing.T) {

	b := val.EmptyImmutableList()

	b = b.Append(val.Boolean(true))
	b = b.Append(val.Long(123))
	b = b.Append(val.Double(-12.34))
	b = b.Append(val.Utf8("text"))
	b = b.Append(val.Raw([]byte{0, 1, 2}, false))

	require.Equal(t, "[true,123,-12.34,\"text\",\"base64,AAEC\"]", val.Jsonify(b))
	require.Equal(t, "95c37bcbc028ae147ae147aea474657874c403000102", val.Hex(b))

	testPackUnpack(t, b)

}
