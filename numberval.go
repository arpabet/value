package genval

import (
	"reflect"
	"strconv"
	"fmt"
)

type NumberValue struct {
	dt 		NumberType
	long 	int64
	double 	float64
}

func Long(val int64) *NumberValue {
	return &NumberValue{
		dt: LONG,
		long: val,
	}
}

func Double(val float64) *NumberValue {
	return &NumberValue{
		dt: DOUBLE,
		double: val,
	}
}

func (n NumberValue) Kind() Kind {
	return NUMBER
}

func (n NumberValue) Class() reflect.Type {
	return reflect.TypeOf((*NumberValue)(nil)).Elem()
}

func (n NumberValue) String() string {
	switch n.dt {
	case LONG:
		return strconv.FormatInt(n.long, 10)
	case DOUBLE:
		return fmt.Sprint(n.double)
	}
	return "nil"
}

func (n NumberValue) Pack(p Packer) {
	switch n.dt {
	case LONG:
		p.PackLong(n.long)
	case DOUBLE:
		p.PackDouble(n.double)
	default:
		p.PackNil()
	}
}

func (n NumberValue) Json() string {
	return n.String()
}

func (n NumberValue) Type() NumberType {
	return n.dt
}

func (n NumberValue) Long() int64 {
	switch n.dt {
	case LONG:
		return n.long
	case DOUBLE:
		return int64(n.double)
	}
	return 0
}

func (n NumberValue) Double() float64 {
	switch n.dt {
	case LONG:
		return float64(n.long)
	case DOUBLE:
		return n.double
	}
	return 0
}

func (n NumberValue) Add(other Number) Number {
	return &n
}

func (n NumberValue) Subtract(other Number) Number {
	return &n
}

