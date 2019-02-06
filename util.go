package genval

import (
	"bytes"
	"encoding/hex"
)

func Hex(val Value) string {

	buf := bytes.Buffer{}
	p := MessagePacker(&buf)
	val.Pack(p)

	return hex.EncodeToString(buf.Bytes())
}
