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