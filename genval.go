package genval

import (
	"reflect"
	"github.com/shvid/genval/genval_impl"
)

/**
	Base interface for all Generic Values

    Author: Alex Shvid
 */


type Kind int

const (
	InvalidVal Kind = iota
	NilVal
	BoolVal
	NumberVal
	StringVal
	TableVal
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
		Pack generic value by using packer
	 */

	Pack(Packer)

	/**
		Converts Generic Value to JSON
	 */

	Json() string

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

type Expression interface {

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
	Parses expression
 */

func Expr(str string) Expression {
	return genval_impl.ParseExpr(str)
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
	InvalidNumber NumberType = iota
	Long
	Double
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
	Bytes
)

type String interface {
	Value

	/**
		Gets string type, that can be UTF8 or Bytes
	 */

	Type() StringType

	/**
		Gets string as utf8 string
	 */

	UTF8() String

	/**
		Gets string as byte array
	 */

	Bytes() []byte

}

/**
	Table interface

    Tables can be Array or Map

    For Array indexes start from 1 and increase sequentially

    Author: Alex Shvid
 */

type TableType int

const (
	InvalidTable TableType = iota
	Array
	Map
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

	GetX(Expression) Value

	/**
		Gets table by the expression
	 */

	GetTableX(Expression) Table

	/**
		Gets boolean value by the expression
	 */

	GetBoolX(Expression) Bool

	/**
		Gets number value by the expression
	 */

	GetNumberX(Expression) Number

	/**
		Gets string value by the expression
	 */

	GetStringX(Expression) String

	/**
		Puts value by the key and returns old one
	 */

	Put(key string, value Value) Value

	/**
		Puts value by the index and returns old one
 	*/

	PutAt(index int, value Value) Value

	/**
		Puts value by the expression and returns old one
 	*/

	PutX(exp Expression, value Value) Value

	/**
		Removes value by the key and returns old one
 	*/

	Remove(string) Value

	/**
		Removes value by the index and returns old one
 	*/

	RemoveAt(int) Value

	/**
		Removes value by the expression and returns old one
 	*/

	RemoveX(Expression) Value

	/**
		Returns Map of entries
	 */

	Map() map[string]Value

	/**
		Returns Array of values
 	*/

	Array() []Value

	/**
		Returns size of the table (number of entries)
	 */

	Size() int

	/**
		Erase all elements in the table
	 */

	Clear()

}


