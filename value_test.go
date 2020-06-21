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
	val "github.com/consensusdb/value"
	"github.com/stretchr/testify/require"
	"bytes"
	"time"
	"runtime"
	"fmt"
)

const numIterations = 10
const numBenchIterations = 10000

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

func TestBenchmark(t *testing.T) {

	m := testCreateMap()

	runtime.GC()
	tnow := time.Now()

	buf := bytes.Buffer{}
	p := val.MessagePacker(&buf)

	for i:=0; i!=numBenchIterations;i++ {
		m.Pack(p)
	}

	encDur := time.Now().Sub(tnow)
	encLen := len(buf.Bytes())
	runtime.GC()

	unpacker := val.MessageUnpacker(buf.Bytes(), false)
	parser := val.MessageParser()

	tnow = time.Now()
	for i:=0; i!=numBenchIterations;i++ {
		val.Parse(unpacker, parser)
	}

	decDur := time.Now().Sub(tnow)

	fmt.Printf("Benchmark %d ops, encode_len=%d, encode_duration=%v, decode_duration=%v\n", numBenchIterations, encLen, encDur, decDur)

	writeBs := int64(time.Second / encDur) * numBenchIterations
	readBs := int64(time.Second / decDur) * numBenchIterations

	fmt.Printf("Throughput write=%v bs, read=%v bs\n", writeBs, readBs)


}

func testCreateMap() val.Table {

	b := val.Map()

	b.Insert(val.Boolean(true))
	b.Insert(val.Long(123))
	b.Insert(val.Double(-12.34))
	b.Insert(val.Utf8("text"))
	b.Insert(val.Raw([]byte{0, 1, 2}, false))

	c := val.Map()
	c.Put("5", val.Long(5))

	b.Put("name", val.Utf8("name"))
	b.Put("123", val.Long(123))
	b.Put("map", c)

	return b
}