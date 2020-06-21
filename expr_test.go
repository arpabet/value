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
)

func TestExpr(t *testing.T) {

	b := val.Expression("")

	require.Equal(t, 1, b.Size())
	require.Equal(t, "", b.GetAt(0))
	require.Equal(t, "", b.String())

	b = val.Expression("name")

	require.Equal(t, 1, b.Size())
	require.Equal(t, "name", b.GetAt(0))
	require.Equal(t, "", b.GetAt(-1))
	require.Equal(t, "", b.GetAt(1))
	require.Equal(t, "name", b.String())

	b = val.Expression("name.first")

	require.Equal(t, 2, b.Size())
	require.Equal(t, "name", b.GetAt(0))
	require.Equal(t, "first", b.GetAt(1))
	require.Equal(t, "", b.GetAt(-1))
	require.Equal(t, "", b.GetAt(2))
	require.Equal(t, "name.first", b.String())

}