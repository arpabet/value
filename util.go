package genval

import (
	"bytes"
	"encoding/hex"
)

func Hex(val Value) string {

	buf := bytes.Buffer{}
	p := NewMessagePacker(&buf)
	val.Pack(p)

	return hex.EncodeToString(buf.Bytes())
}
