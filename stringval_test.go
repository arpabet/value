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
	"strconv"
)

var testStrings = map[string]string {

	"": "a0",
	"test": "a474657374",
	"hello": "a568656c6c6f",
}

func TestUtf8String(t *testing.T) {

	for str, hex := range testStrings {

		b := genval.Utf8(str)

		require.Equal(t, genval.STRING, b.Kind())
		require.Equal(t, genval.UTF8, b.Type())
		require.Equal(t, "genval.stringValue", b.Class().String())
		require.Equal(t, hex, genval.Hex(b))
		require.Equal(t, "\""+ str + "\"", genval.Json(b))
		require.Equal(t, str, b.String())

	}

}

func TestJsonString(t *testing.T) {

	src := "json\"val\"json"

	s := genval.Utf8(src)

	require.Equal(t, src, s.String())
	require.Equal(t, "\"json\\\"val\\\"json\"", genval.Json(s))

	actual, _ := strconv.Unquote(genval.Json(s))
	require.Equal(t, src, actual)

}

func TestRawString(t *testing.T) {

	raw := []byte { 0, 1, 2, 3, 4, 5 }
	s := genval.Raw(raw, false)

	require.Equal(t, genval.STRING, s.Kind())
	require.Equal(t, genval.RAW, s.Type())
	require.Equal(t, genval.Base64Prefix + "AAECAwQF", s.String())
	require.Equal(t, "\"" + genval.Base64Prefix + "AAECAwQF\"", genval.Json(s))
	require.Equal(t, "c406000102030405", genval.Hex(s))
	require.Equal(t, 0, bytes.Compare(raw, s.Raw()))

	actual := genval.ParseString(s.String())
	//equire.Equal(t, 0, bytes.Compare(s.Raw(), actual.Raw()))
	require.Equal(t, s.Raw(), actual.Raw())
}