package genval_impl

import (
	"reflect"
	"io"
	"github.com/shvid/genval"
)

type BoolVal bool

func (b BoolVal) Kind() genval.Kind {
	return genval.BoolVal
}

func (b BoolVal) Class() reflect.Type {
	return reflect.TypeOf(BoolVal(false))
}

func (b BoolVal) String() string {
	if b {
		return "true"
	} else {
		return "false"
	}
}

func (b BoolVal) ByteString() []byte {
	if b {
		return []byte { mpTrue }
	} else {
		return []byte { mpFalse }
	}
}

func (b BoolVal) WriteTo(writer io.Writer) {
	if b {
		writer.Write([]byte { mpTrue })
	} else {
		writer.Write([]byte { mpFalse })
	}
}

func (b BoolVal) Json() string {
	return b.String()
}

func (b BoolVal) Boolean() bool {
	return bool(b)
}


