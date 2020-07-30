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
	"encoding"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"math/big"
	"reflect"
	"strings"
)

/**
	Base interface for all values

    Author: Alex Shvid
 */


type Kind int

const (
	INVALID Kind = iota
	BOOL
	NUMBER
	STRING
	LIST
	MAP
	UNKNOWN
)

type Value interface {
	fmt.Stringer
	json.Marshaler
	encoding.BinaryMarshaler

	/**
		Gets value kind type for easy reflection
	 */

	Kind() Kind

	/**
		Gets reflection type for easy reflection operations
 	*/

	Class() reflect.Type

	/**
		Gets underline object
	 */

	Object() interface{}

	/**
		Pack generic value by using packer, must not be nil
	 */

	Pack(Packer)

	/**
		Converts Generic Value to JSON
	 */

	PrintJSON(out *strings.Builder)

	/**
		Check if values are equal, nil friendly function
	 */

	Equal(Value) bool

}


/**
	Boolean interface

    Author: Alex Shvid
 */

type Bool interface {
	Value

	/**
		Gets payload as boolean
	 */

	Boolean() bool
}

/**
	Number interface

    Numbers can be int64 and double

    Author: Alex Shvid
 */

type NumberType int

const (
	InvalidNumber 	NumberType = iota
	LONG
	DOUBLE
	BIGINT
	DECIMAL
)

func (t NumberType) String() string {
	switch t {
	case InvalidNumber:
		return "invalid"
	case LONG:
		return "long"
	case DOUBLE:
		return "double"
	case BIGINT:
		return "bigint"
	case DECIMAL:
		return "decimal"
	default:
		return "unknown"
	}
}

type Number interface {
	Value

	/**
		Gets number type, supported only long and double
	 */

	Type() NumberType

	/**
		Check if number is not a number
	 */

	IsNaN() bool

	/**
		Gets number as long
	 */

	Long() int64

	/**
		Gets number as double
	 */

	Double() float64

	/**
		Gets number as BigInt
	 */

	BigInt() *big.Int

	/**
		Gets number as Decimal
	 */

	Decimal() decimal.Decimal

	/**
		Adds this number and other one and return a new one
	 */

	Add(Number) Number

	/**
		Subtracts from this number the other one and return a new one
	 */

	Subtract(Number) Number

}

/**
	String interface

    Strings can be UTF-8 and ByteStrings

    Author: Alex Shvid
 */

type StringType int

const (
	InvalidString StringType = iota
	UTF8
	RAW
)

func (t StringType) String() string {
	switch t {
	case InvalidString:
		return "invalid"
	case UTF8:
		return "utf8"
	case RAW:
		return "raw"
	default:
		return "unknown"
	}
}

type String interface {
	Value

	/**
		Gets string type, that can be UTF8 or Bytes
	 */

	Type() StringType

	/**
		Length of the string
	 */

	Len() int

	/**
		Gets string as utf8 string
	 */

	Utf8() string

	/**
		Gets string as byte array
	 */

	Raw() []byte

}

type Extension interface {
	Value

	Native() []byte
}


type ListItem interface {

	/*
		Index in the array where item is located
	 */
	Key() int

	Value()  Value

	Equal(ListItem) bool

}

type MapEntry interface {

	Key() string

	Value() Value

	Equal(MapEntry) bool

}


type Collection interface {

	/**
		Get entries of all element like in Map
	*/

	Entries()  []MapEntry

}

type List interface {
	Value
	Collection

	/**
		List items
	 */

	Items() []ListItem

	/**
		List values
	*/

	Values() []Value

	/**
		Length of the list
	*/

	Len() int

	/**
		Gets value by the index

	    return value or nil
	*/

	GetAt(int) Value

	/**
		Gets boolean value by the index

	    return value or nil
	*/

	GetBoolAt(int) Bool

	/**
		Gets number value by the index

	    return value or nil
	*/

	GetNumberAt(int) Number

	/**
		Gets string value by the index

	    return value or nil
	*/

	GetStringAt(int) String

	/**
		Gets list by the index

	    return value or nil
	*/

	GetListAt(int) List

	/**
		Gets map by the index

	    return value or nil
	*/

	GetMapAt(int) Map

	/**
		Sets value to the list at position i
	*/

	PutAt(int, Value) List

	/**
		Adds value to the list at position i by shifting to left
	*/

	InsertAt(int, Value) List

	/**
		Adds value to the list, same as Add or Insert
	*/

	Append(Value) List

	/**
		Removes value by the index
	*/

	RemoveAt(int) List

}


type Map interface {
	Value
	Collection

	/**
		Construct standard Hash Map
	 */

	HashMap() map[string]Value

	/**
		List keys
	*/

	Keys() []string

	/**
		List values
	*/

	Values() []Value

	/**
		Length of the map
	*/

	Len() int

	/**
		Gets value by the key

	    return (value or nil, true) or (nil, false)
	*/

	Get(string) (Value, bool)


	/**
		Gets boolean value by the key

	    return value or nil
	*/

	GetBool(string) Bool

	/**
		Gets number value by the key

	    return value or nil
	*/

	GetNumber(string) Number

	/**
		Gets string value by the key

		return value or nil
	*/

	GetString(string) String

	/**
		Gets list by the key

	    return value or nil
	*/

	GetList(string) List

	/**
		Gets list by the key

	    return value or nil
	*/

	GetMap(string) Map

	/**
		Inserts value at specific key, do not remove doubles
	*/

	Insert(key string, value Value) Map

	/**
		Puts value by the key, replaces if it exist
	*/

	Put(key string, value Value) Map

	/**
		Removes value by the key
	*/

	Remove(string) Map

}





