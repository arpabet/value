package genval

import (
	"bytes"
	"encoding/hex"
)

func Pack(val Value) []byte {

	buf := bytes.Buffer{}
	p := MessagePacker(&buf)
	val.Pack(p)

	return buf.Bytes()
}


func Hex(val Value) string {
	return hex.EncodeToString(Pack(val))
}
