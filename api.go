/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
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
*/

var AllowFastAppends = true

type Kind int

const (
	INVALID Kind = iota
	NULL
	BOOL
	NUMBER
	STRING
	LIST
	MAP
	UNKNOWN
)

func (k Kind) String() string {
	switch k {
	case INVALID:
		return "INVALID"
	case BOOL:
		return "BOOL"
	case NUMBER:
		return "NUMBER"
	case STRING:
		return "STRING"
	case LIST:
		return "LIST"
	case MAP:
		return "MAP"
	case UNKNOWN:
		return "UNKNOWN"
	default:
		return "DEFAULT"
	}
}

var ValueClass = reflect.TypeOf((*Value)(nil)).Elem()

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


type Updater interface {

	/**
	In-pace updater of the value
	 */

	Update(old Value) (new Value)

}

/**
Boolean interface

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

*/

type NumberType int

const (
	InvalidNumber NumberType = iota
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

	/**
	Gets serialized extension in MsgPack
	 */

	Native() []byte
}

type ListItem interface {

	/**
	Index in the array where item is located
	*/
	Key() int

	/**
	Value of the cell
	 */

	Value() Value

	/**
	Updates value in-place

	Returns true if value was updated
	*/

	Update(Updater) bool

	/**
	Checks if key and value are the same
	*/

	Equal(ListItem) bool
}

type MapEntry interface {

	/**
	Key of the map always a string
	 */

	Key() string

	/**
	Value if the map can be any object with Value interface
	 */

	Value() Value

	/**
	Updates value in-place

	Returns true if value was updated
	*/

	Update(Updater) bool

	/**
	Checks if key and value are the same
	 */

	Equal(MapEntry) bool
}

type Collection interface {

	/**
	Length of the Collection
	*/

	Len() int

	/**
	Get entries of all element like in Map
	*/

	Entries() []MapEntry
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
	Updates value to the list at position i, does not work in immutable maps

	Returns true if value was updated
	*/

	UpdateAt(int, Updater) bool

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

	/**
		Gets all values with the same key

	    return value array or nil
	*/

	Select(int) []Value

	/**
	Insert all values with the same key
	*/

	InsertAll(int, []Value) List

	/**
	Delete all values with the same key
	*/

	DeleteAll(int) List
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
		Gets value by the key

	    return Value or Null
	*/

	Get(string) Value

	/**
		Gets boolean value by the key

	    return Bool or False
	*/

	GetBool(string) Bool

	/**
		Gets number value by the key

	    return Number or Zero
	*/

	GetNumber(string) Number

	/**
	Gets string value by the key

	return String or Str("")
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
	Updates value in place, does not work for immutable maps

	Returns true if value was updated
	*/

	Update(key string, updater Updater) bool

	/**
	Removes value by the key
	*/

	Remove(string) Map

	/**
		Gets all values with the same key

	    return value array or nil
	*/

	Select(string) []Value

	/**
	Insert all values with the same key
	*/

	InsertAll(string, []Value) Map

	/**
	Delete all values with the same key
	*/

	DeleteAll(string) Map
}
