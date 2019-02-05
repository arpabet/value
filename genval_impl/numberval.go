package genval_impl

import (
	"github.com/shvid/genval"
	"reflect"
	"strconv"
	"fmt"
)

type NumberVal struct {
	dt 		genval.NumberType
	long 	int64
	double 	float64
}

func (n NumberVal) Kind() genval.Kind {
	return genval.NumberVal
}

func (n NumberVal) Class() reflect.Type {
	return reflect.TypeOf((*NumberVal)(nil)).Elem()
}

func (n NumberVal) String() string {
	switch n.dt {
	case genval.Long:
		return strconv.FormatInt(n.long, 10)
	case genval.Double:
		return fmt.Sprint(n.double)
	}
	return "nil"
}

func (n NumberVal) Pack(p genval.Packer) {
	switch n.dt {
	case genval.Long:
		p.PackLong(n.long)
	case genval.Double:
		p.PackDouble(n.double)
	default:
		p.PackNil()
	}
}

func (n NumberVal) Json() string {
	return n.String()
}

func (n NumberVal) Type() genval.NumberType {
	return n.dt
}

func (n NumberVal) Long() int64 {
	switch n.dt {
	case genval.Long:
		return n.long
	case genval.Double:
		return int64(n.double)
	}
	return 0
}

func (n NumberVal) Double() float64 {
	switch n.dt {
	case genval.Long:
		return float64(n.long)
	case genval.Double:
		return n.double
	}
	return 0
}

func (n NumberVal) Add(other genval.Number) genval.Number {
	return &n
}

func (n NumberVal) Subtract(other genval.Number) genval.Number {
	return &n
}

