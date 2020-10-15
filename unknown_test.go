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
	val "arpabet.pkg.is/value"
	"bytes"
)

func TestUnknown(t *testing.T) {

	tagAndData := []byte { byte(val.MaxExt),  1 }

	v := val.Unknown(tagAndData)

	require.Equal(t, val.UNKNOWN, v.Kind())
	require.Equal(t, val.UnknownPrefix+ val.Base64Prefix + "AwE", v.String())
	require.Equal(t, "\"" + v.String() + "\"", val.Jsonify(v))
	require.Equal(t, "d40301", val.Hex(v))
	require.Equal(t, 0, bytes.Compare(tagAndData, v.Native()))

	mp, err := val.Pack(v)
	require.Nil(t, err)

	a, err := val.Unpack(mp, false)
	require.Nil(t, err)

	require.Equal(t, val.UNKNOWN, v.Kind())
	require.Equal(t, 0, bytes.Compare(tagAndData, a.(val.Extension).Native()))

}