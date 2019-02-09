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
	"encoding/json"
)

func TestBool(t *testing.T) {

	b := genval.Boolean(true)

	require.Equal(t, genval.BOOL, b.Kind())
	require.Equal(t, "genval.boolValue", b.Class().String())
	require.Equal(t, "c3", genval.Hex(b))
	require.Equal(t, "true", genval.Json(b))
	require.Equal(t, "true", b.String())

	require.Equal(t, true, genval.ParseBoolean("t").Boolean())
	require.Equal(t, true, genval.ParseBoolean("true").Boolean())
	require.Equal(t, true, genval.ParseBoolean("True").Boolean())

	b = genval.Boolean(false)
	require.Equal(t, "c2", genval.Hex(b))
	require.Equal(t, "false", genval.Json(b))
	require.Equal(t, "false", b.String())

	require.Equal(t, false, genval.ParseBoolean("f").Boolean())
	require.Equal(t, false, genval.ParseBoolean("false").Boolean())
	require.Equal(t, false, genval.ParseBoolean("False").Boolean())
	require.Equal(t, false, genval.ParseBoolean("").Boolean())
	require.Equal(t, false, genval.ParseBoolean("any_value").Boolean())

}

type testBoolStruct struct {
	B genval.Bool
}

func TestBoolMarshal(t *testing.T) {

	b := genval.Boolean(true)

	j, _ := b.MarshalJSON()
	require.Equal(t, []byte("true"), j)

	bin, _ := b.MarshalBinary()
	require.Equal(t, []byte{0xc3}, bin)

	b = genval.Boolean(false)

	j, _ = b.MarshalJSON()
	require.Equal(t, []byte("false"), j)

	bin, _ = b.MarshalBinary()
	require.Equal(t, []byte{0xc2}, bin)

	s := &testBoolStruct{genval.Boolean(true)}

	j, _ = json.Marshal(s)
	require.Equal(t, "{\"B\":true}", string(j))

}

func TestPackBool(t *testing.T) {

	b := genval.Boolean(true)

	mp, _ := genval.Pack(b)

	c, err := genval.Unpack(mp, false)
	if err != nil {
		t.Errorf("unpack fail %v", err)
	}

	require.True(t, b.Equal(c))

}