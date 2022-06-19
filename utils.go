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

package value

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/nacl/box"
	"io"
	"strings"
)



func Pack(val Value) ([]byte, error) {
	buf := bytes.Buffer{}
	p := MessagePacker(&buf)
	if val != nil {
		val.Pack(p)
	} else {
		p.PackNil()
	}
	return buf.Bytes(), p.Error()
}

func Unpack(buf []byte, copy bool) (Value, error) {
	unpacker := MessageUnpacker(buf, copy)
	parser := MessageParser()
	return Parse(unpacker, parser)
}

func Read(r io.Reader) (Value, error) {
	unpacker := MessageReader(r)
	parser := MessageParser()
	return Parse(unpacker, parser)
}

func Write(w io.Writer, val Value) error {
	p := MessagePacker(w)
	val.Pack(p)
	return p.Error()
}

func Hex(val Value) string {
	mp, _ := Pack(val)
	return hex.EncodeToString(mp)
}

func Jsonify(val Value) string {
	var out strings.Builder
	if val != nil {
		val.PrintJSON(&out)
	} else {
		out.WriteString("null")
	}
	return out.String()
}

func Hash(val Value, hash crypto.Hash) ([]byte, error) {
	data, err := Pack(val)
	if err != nil {
		return nil, err
	}
	return hash.New().Sum(data), nil
}

// use box.GenerateKey(rand.Reader) to get keys
func Seal(val Value, recipientPublicKey, senderPrivateKey *[32]byte) ([]byte, error){
	var nonce [24]byte
	_, err := io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		return nil, err
	}
	unencrypted, err := Pack(val)
	if err != nil {
		return nil, err
	}
	encrypted := box.Seal(nonce[:], unencrypted, &nonce, recipientPublicKey, senderPrivateKey)
	return encrypted, nil
}

var ErrUnseal = errors.New("unseal error")

func Unseal(encrypted []byte, senderPublicKey, recipientPrivateKey *[32]byte) (Value, error) {
	var decryptNonce [24]byte
	copy(decryptNonce[:], encrypted[:24])
	decrypted, ok := box.Open(nil, encrypted[24:], &decryptNonce, senderPublicKey, recipientPrivateKey)
	if !ok {
		return nil, ErrUnseal
	}
	return Unpack(decrypted, false)
}

func Equal(left Value, right Value) bool {
	if left == nil {
		return right == nil
	}
	return left.Equal(right)
}

func Parse(unpacker Unpacker, parser Parser) (Value, error) {
	return doParse(unpacker, parser)
}

func WriteStream(w io.Writer, valueC <-chan Value) error {

	p := MessagePacker(w)

	for p.Error() == nil {
		val, ok := <- valueC

		if !ok {
			break
		}

		val.Pack(p)

	}

	return p.Error()
}

func ReadStream(r io.Reader, out chan<- Value) error {

	defer close(out)

	unpacker := MessageReader(r)
	parser := MessageParser()

	for {

		value, err := doParse(unpacker, parser)
		if err != nil {
			return err
		}

		out <- value
	}

	return nil
}

