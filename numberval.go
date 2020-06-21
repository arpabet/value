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

package value

import (
	"reflect"
	"strconv"
	"fmt"
	"math"
	"strings"
	"github.com/pkg/errors"
)

const (
	precisionLevel = 0.00001
)

type numberValue struct {
	dt 		NumberType
	long 	int64
	double 	float64
}

func (n numberValue) Equal(val Value) bool {
	if val == nil || val.Kind() != NUMBER {
		return false
	}
	o := val.(*numberValue)
	if n.dt != o.dt {
		return false
	}
	switch n.dt {
	case LONG:
		return n.long == o.long
	case DOUBLE:
		if math.IsNaN(n.double) {
			return math.IsNaN(o.double)
		} else {
			return math.Abs(n.double - o.double) < precisionLevel
		}
	}
	return false
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

func Nan() *numberValue {
	return &numberValue{
		dt: DOUBLE,
		double: math.NaN(),
	}
}

func ParseNumber(str string) *numberValue {

	if len(str) == 0 {
		return Long(0)
	}

	long, err := strconv.ParseInt(str, 10, 64)
	if err == nil {
		return Long(long)
	} else {
		double, err := strconv.ParseFloat(str, 64)
		if err == nil {
			return Double(double)
		}
	}

	return Nan()

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
	return ""
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

func (n numberValue) PrintJSON(out *strings.Builder) {
	switch n.dt {
	case LONG:
		out.WriteString(strconv.FormatInt(n.long, 10))
	case DOUBLE:
		if math.IsNaN(n.double) {
			out.WriteString("null")
		} else {
			out.WriteString(fmt.Sprint(n.double))
		}
	default:
		out.WriteString("null")
	}
}

func (n numberValue) MarshalJSON() ([]byte, error) {
	switch n.dt {
	case LONG:
		return []byte(strconv.FormatInt(n.long, 10)), nil
	case DOUBLE:
		if math.IsNaN(n.double) {
			return []byte("null"), nil
		} else {
			return []byte(fmt.Sprint(n.double)), nil
		}
	default:
		return nil, errors.New("unknown data type")
	}
}

func (n numberValue) MarshalBinary() ([]byte, error) {
	m := new(messageWriter)  // must be in heap
	switch n.dt {
	case LONG:
		return m.WriteLong(n.long), nil
	case DOUBLE:
		return m.WriteDouble(n.double), nil
	default:
		return nil, errors.New("unknown data type")
	}
}

func (n numberValue) Type() NumberType {
	return n.dt
}

func (n numberValue) Long() int64 {
	switch n.dt {
	case LONG:
		return n.long
	case DOUBLE:
		if math.IsNaN(n.double) {
			return 0
		} else {
			return int64(n.double)
		}
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
	return math.NaN()
}

func (n numberValue) Add(other Number) Number {
	switch n.dt {
	case LONG:
		return Long(n.long + other.Long())
	case DOUBLE:
		right := other.Double()
		if math.IsNaN(n.double) || math.IsNaN(right) {
			return Nan()
		}
		return Double(n.double + right)
	}
	return Nan()
}

func (n numberValue) Subtract(other Number) Number {
	switch n.dt {
	case LONG:
		return Long(n.long - other.Long())
	case DOUBLE:
		right := other.Double()
		if math.IsNaN(n.double) || math.IsNaN(right) {
			return Nan()
		}
		return Double(n.double - right)
	}
	return Nan()
}

