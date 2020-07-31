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
	"testing"
	"github.com/stretchr/testify/require"
	val "github.com/consensusdb/value"
	"encoding/json"
)

/**
	@author Alex Shvid
*/

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