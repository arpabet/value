/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */



package value_test

import (
	"testing"
	"github.com/stretchr/testify/require"
	val "go.arpabet.com/value"
	"encoding/json"
)

func TestBool(t *testing.T) {

	b := val.Boolean(true)

	require.Equal(t, val.BOOL, b.Kind())
	require.Equal(t, "value.boolValue", b.Class().String())
	require.Equal(t, "c3", val.Hex(b))
	require.Equal(t, "true", val.Jsonify(b))
	require.Equal(t, "true", b.String())

	require.Equal(t, true, val.ParseBoolean("t").Boolean())
	require.Equal(t, true, val.ParseBoolean("true").Boolean())
	require.Equal(t, true, val.ParseBoolean("True").Boolean())

	b = val.Boolean(false)
	require.Equal(t, "c2", val.Hex(b))
	require.Equal(t, "false", val.Jsonify(b))
	require.Equal(t, "false", b.String())

	require.Equal(t, false, val.ParseBoolean("f").Boolean())
	require.Equal(t, false, val.ParseBoolean("false").Boolean())
	require.Equal(t, false, val.ParseBoolean("False").Boolean())
	require.Equal(t, false, val.ParseBoolean("").Boolean())
	require.Equal(t, false, val.ParseBoolean("any_value").Boolean())

}

type testBoolStruct struct {
	B val.Bool
}

func TestBoolMarshal(t *testing.T) {

	b := val.Boolean(true)

	j, _ := b.MarshalJSON()
	require.Equal(t, []byte("true"), j)

	bin, _ := b.MarshalBinary()
	require.Equal(t, []byte{0xc3}, bin)

	b = val.Boolean(false)

	j, _ = b.MarshalJSON()
	require.Equal(t, []byte("false"), j)

	bin, _ = b.MarshalBinary()
	require.Equal(t, []byte{0xc2}, bin)

	s := &testBoolStruct{val.Boolean(true)}

	j, _ = json.Marshal(s)
	require.Equal(t, "{\"B\":true}", string(j))

}

func TestPackBool(t *testing.T) {

	b := val.Boolean(true)
	testPackUnpack(t, b)

	b = val.Boolean(false)
	testPackUnpack(t, b)

}