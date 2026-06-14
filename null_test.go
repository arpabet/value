/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */



package value_test

import (
	"testing"
	"github.com/stretchr/testify/require"
	val "go.arpabet.com/value"
	"encoding/json"
)

func TestNull(t *testing.T) {

	b := val.Null

	require.Equal(t, val.NULL, b.Kind())
	require.Equal(t, "value.nullValue", b.Class().String())
	require.Equal(t, "c0", val.Hex(b))
	require.Equal(t, "null", val.Jsonify(b))
	require.Equal(t, "null", b.String())

	require.Equal(t, val.STRING, val.ParseNull("something").Kind())
	require.Equal(t, val.NULL, val.ParseNull("null").Kind())

}

type testNullStruct struct {
	B val.Value
}

func TestNullMarshal(t *testing.T) {

	b := val.Null

	j, _ := b.MarshalJSON()
	require.Equal(t, []byte("null"), j)

	bin, _ := b.MarshalBinary()
	require.Equal(t, []byte{0xc0}, bin)

	s := &testNullStruct{val.Nil() }

	j, _ = json.Marshal(s)
	require.Equal(t, "{\"B\":null}", string(j))

}

func TestPackNull(t *testing.T) {

	b := val.Null
	testPackUnpack(t, b)

}