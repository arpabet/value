package genval

import (
	"reflect"
	"strconv"
	"fmt"
)

type numberValue struct {
	dt 		NumberType
	long 	int64
	double 	float64
}

func Long(val int64) *numberValue {
	return &numberValue{
		dt: LONG,
		long: val,
	}
}

func Double(val float64) *numberValue {
	return &numberValue{
		dt: DOUBLE,
		double: val,
	}
}

func (n numberValue) Kind() Kind {
	return NUMBER
}

func (n numberValue) Class() reflect.Type {
	return reflect.TypeOf((*numberValue)(nil)).Elem()
}

func (n numberValue) String() string {
	switch n.dt {
	case LONG:
		return strconv.FormatInt(n.long, 10)
	case DOUBLE:
		return fmt.Sprint(n.double)
	}
	return "nil"
}

func (n numberValue) Pack(p Packer) {
	switch n.dt {
	case LONG:
		p.PackLong(n.long)
	case DOUBLE:
		p.PackDouble(n.double)
	default:
		p.PackNil()
	}
}

func (n numberValue) Json() string {
	return n.String()
}

func (n numberValue) Type() NumberType {
	return n.dt
}

func (n numberValue) Long() int64 {
	switch n.dt {
	case LONG:
		return n.long
	case DOUBLE:
		return int64(n.double)
	}
	return 0
}

func (n numberValue) Double() float64 {
	switch n.dt {
	case LONG:
		return float64(n.long)
	case DOUBLE:
		return n.double
	}
	return 0
}

func (n numberValue) Add(other Number) Number {
	return &n
}

func (n numberValue) Subtract(other Number) Number {
	return &n
}

