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
	"bytes"
)

func TestUnknown(t *testing.T) {

	mp := []byte { 0xd4,  1 }

	v := val.Unknown(mp, nil)

	require.Equal(t, val.UNKNOWN, v.Kind())
	require.Equal(t, val.DataPrefix + val.Base64Prefix + "1AE", v.String())
	require.Equal(t, "\"" + v.String() + "\"", val.Json(v))
	require.Equal(t, "d401", val.Hex(v))
	require.Equal(t, 0, bytes.Compare(mp, v.Packed()))


	mp = []byte { 0xc7 }
	p := []byte { 0, 1 }

	v = val.Unknown(mp, p)

	require.Equal(t, val.UNKNOWN, v.Kind())
	require.Equal(t, val.DataPrefix + val.Base64Prefix + "xwAB", v.String())
	require.Equal(t, "\"" + v.String() + "\"", val.Json(v))
	require.Equal(t, "c70001", val.Hex(v))
	require.Equal(t, 0, bytes.Compare([]byte { 0xc7, 0, 1 }, v.Packed()))

}