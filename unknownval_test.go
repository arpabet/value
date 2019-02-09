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
	"bytes"
)

func TestUnknown(t *testing.T) {

	mp := []byte { 0xd4,  1 }

	v := genval.Unknown(mp, nil)

	require.Equal(t, genval.UNKNOWN, v.Kind())
	require.Equal(t, genval.DataPrefix + genval.Base64Prefix + "1AE", v.String())
	require.Equal(t, "\"" + v.String() + "\"", genval.Json(v))
	require.Equal(t, "d401", genval.Hex(v))
	require.Equal(t, 0, bytes.Compare(mp, v.Packed()))


	mp = []byte { 0xc7 }
	p := []byte { 0, 1 }

	v = genval.Unknown(mp, p)

	require.Equal(t, genval.UNKNOWN, v.Kind())
	require.Equal(t, genval.DataPrefix + genval.Base64Prefix + "xwAB", v.String())
	require.Equal(t, "\"" + v.String() + "\"", genval.Json(v))
	require.Equal(t, "c70001", genval.Hex(v))
	require.Equal(t, 0, bytes.Compare([]byte { 0xc7, 0, 1 }, v.Packed()))

}