/**
    Copyright (c) 2020-2022 Arpabet, Inc.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in
	all copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
	THE SOFTWARE.
*/

package value_test


import (
	"testing"
	val "go.arpabet.com/value"
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

func testCreateMap() val.Map {

	a := val.EmptySparseList()
	a = a.InsertAt(1, val.Boolean(true))
	a = a.InsertAt(0, val.Long(123))
	a = a.InsertAt(3, val.Double(-12.34))
	a = a.InsertAt(1, val.Utf8("text"))
	a = a.InsertAt(5, val.Raw([]byte{0, 1, 2}, false))

	b := val.EmptyList()
	b = b.Append(val.Boolean(true))
	b = b.Append(val.Long(123))
	b = b.Append(val.Double(-12.34))
	b = b.Append(val.Utf8("text"))
	b = b.Append(val.Raw([]byte{0, 1, 2}, false))

	c := val.EmptyMap()
	c = c.Put("5", val.Long(5))

	c = c.Put("name", val.Utf8("name"))
	c = c.Put("123", val.Long(123))
	c = c.Put("list", b)

	return c
}

func TestPackNil(t *testing.T) {

	data, err := val.Pack(nil)
	require.Nil(t, err)

	require.Equal(t, []byte{0xc0}, data)

	actual, err := val.Unpack(data, false)
	require.Nil(t, err)
	require.Nil(t, actual)


}