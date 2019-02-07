package genval

import (
	"reflect"
)

/**
	Base interface for all Generic Values

    Author: Alex Shvid
 */


type Kind int

const (
	INVALID Kind = iota
	BOOL
	NUMBER
	STRING
	TABLE
)

type Value interface {

	/**
		Gets value kind type for easy reflection
	 */

	Kind() Kind

	/**
		Gets reflection type for easy reflection operations
 	*/

	Class() reflect.Type

	/**
		Converts Generic Value to String
	 */

	String() string

	/**
		Pack generic value by using packer, must not be nil
	 */

	Pack(Packer)

	/**
		Converts Generic Value to JSON
	 */

	Json() string

	/**
		Check if values are equal, nil friendly function
	 */

	Equal(Value) bool
}

/**
	Base interface for the packing values
 */

type Packer interface {

	PackNil()

	PackBool(bool)

	PackLong(int64)

	PackDouble(float64)

	PackString(string)

	PackBytes([]byte)

	Error() error

}

/**
	Expression

    Author: Alex Shvid
 */

type Expr interface {

	/**
		Returns true if expression is empty
	 */

	Empty() bool

	/**
		Returns number of tokens in expression
	 */

	Size() int

	/**
		Gets token at the index
	 */

	GetAt(int) string

	/**
		Gets the whole path of the value
	 */

	GetPath() []string

	/**
		Outputs expression as a string
	 */

	String() string

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
)

type Number interface {
	Value

	/**
		Gets number type, supported only long and double
	 */

	Type() NumberType

	/**
		Gets number as long
	 */

	Long() int64

	/**
		Gets number as double
	 */

	Double() float64

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

/**
	Table interface

    Tables can be List or Map

    For List indexes start from 1 and increase sequentially

    Author: Alex Shvid
 */

type TableType int

const (
	InvalidTable TableType = iota
	LIST
	MAP
)

type Table interface {
	Value

	/**
		Gets type of the table
	 */

	Type()  TableType

	/**
		Gets value by the key
	 */

	Get(string) Value

	/**
		Gets table by the key
	 */

	GetTable(string) Table

	/**
		Gets boolean value by the key
	 */

	GetBool(string) Bool

	/**
		Gets number value by the key
	 */

	GetNumber(string) Number

	/**
		Gets string value by the key
	 */

	GetString(string) String

	/**
		Gets value by the index
 	*/

	GetAt(int) Value

	/**
		Gets table by the index
	 */

	GetTableAt(int) Table

	/**
		Gets boolean value by the index
	 */

	GetBoolAt(int) Bool

	/**
		Gets number value by the index
	 */

	GetNumberAt(int) Number

	/**
		Gets string value by the index
	 */

	GetStringAt(int) String

	/**
		Gets value by the expression
 	*/

	GetExp(Expr) Value

	/**
		Gets table by the expression
	 */

	GetTableExp(Expr) Table

	/**
		Gets boolean value by the expression
	 */

	GetBoolExp(Expr) Bool

	/**
		Gets number value by the expression
	 */

	GetNumberExp(Expr) Number

	/**
		Gets string value by the expression
	 */

	GetStringExp(Expr) String

	/**
		Adds value to table (list), equivalent of Put(MaxIndex()+1, value)
	 */

	Insert(Value)

	/**
		Puts value by the key and returns old one
	 */

	Put(key string, value Value)

	/**
		Puts value by the index and returns old one
 	*/

	PutAt(index int, value Value)

	/**
		Puts value by the expression and returns old one
 	*/

	PutExp(exp Expr, value Value)

	/**
		Removes value by the key and returns old one
 	*/

	Remove(string)

	/**
		Removes value by the index and returns old one
 	*/

	RemoveAt(int)

	/**
		Removes value by the expression and returns old one
 	*/

	RemoveExp(Expr)

	/**
		Returns key-value pairs in map
	 */

	Map() map[string]Value

	/**
		Returns values as a slice
 	*/

	List() []Value

	/**
		Gets sorted indexes and keys
	 */

	Keys() []string

	/**
		Gets sorted indexes only
	 */

	Indexes() []int

	/**
		Returns max index for list or 0
	 */

	MaxIndex() int

	/**
		Returns size of the table (number of entries)
	 */

	Size() int

	/**
		Erase all elements in the table
	 */

	Clear()

}


