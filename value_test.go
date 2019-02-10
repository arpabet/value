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

package val_test

import (
	"testing"
	"github.com/shvid/val"
	"github.com/stretchr/testify/require"
	"bytes"
	"time"
)

const numIterations = 10

func testPackUnpack(t *testing.T, v val.Value) {

	mp, _ := val.Pack(v)

	c, err := val.Unpack(mp, false)
	if err != nil {
		t.Errorf("unpack fail %v", err)
	}

	require.True(t, v.Equal(c))

}

func TestStream(t *testing.T) {

	m := val.Utf8("value")

	buf := bytes.Buffer{}

	valueC := make(chan val.Value)
	go val.WriteStream(&buf, valueC)

	for i:=0; i!=numIterations; i++ {
		valueC <- m
	}

	close(valueC)
	time.Sleep(time.Millisecond)

	valueC = make(chan val.Value)
	go val.ReadStream(&buf, valueC)

	cnt := 0
	for {
		val, ok := <- valueC

		if !ok {
			break
		}

		require.True(t, m.Equal(val))

		cnt = cnt + 1
	}

	require.Equal(t, numIterations, cnt)


}