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

package value_test

import (
	"encoding/hex"
	"github.com/arpabet/value"
	"github.com/stretchr/testify/require"
	"testing"
)

/**
	@author Alex Shvid
*/

type Inner struct {

	value.String	 ` tag:"1"  `

}

type Example struct {

	BoolField       value.Bool      `tag:"1"`
	NumberField     value.Number	`tag:"2"`
	StringField     value.String	`tag:"3"`
	ListField       value.List      `tag:"4"`
	MapField        value.Map       `tag:"5"`
	InnerField      *Inner          `tag:"100"`

}

func TestNilStruct(t *testing.T) {

	blob, err := value.PackStruct(nil)
	require.Nil(t, err)
	require.Equal(t,"c0", hex.EncodeToString(blob))

}

func TestEmptyStruct(t *testing.T) {

	var s Example
	blob, err := value.PackStruct(&s)
	require.Nil(t, err)
	require.Equal(t,"80", hex.EncodeToString(blob))

}

func TestStruct(t *testing.T) {

	s := Example{
		BoolField: value.True,
		NumberField: value.Long(123),
		StringField: value.Utf8("test"),
		ListField: value.EmptyList(),
		MapField: value.EmptyMap(),
		InnerField: &Inner {
			String: value.Utf8("inner"),
		},
	}

	blob, err := value.PackStruct(&s)
	require.Nil(t, err)


	var d Example
	err = value.UnpackStruct(blob, &d, false)
	require.Nil(t, err)

	require.True(t, s.BoolField.Equal(d.BoolField))
	require.True(t, s.NumberField.Equal(d.NumberField))
	require.True(t, s.StringField.Equal(d.StringField))
	require.True(t, s.ListField.Equal(d.ListField))
	require.True(t, s.MapField.Equal(d.MapField))
	require.NotNil(t, d.InnerField)
	require.True(t, s.InnerField.String.Equal(d.InnerField.String))


	obj, err := value.Unpack(blob, false)
	require.Nil(t, err)
	require.Equal(t, value.LIST, obj.Kind())
	list := obj.(value.List)
	require.Equal(t, 101, list.Len())

	require.True(t, s.BoolField.Equal(list.GetAt(1)))
	require.True(t, s.NumberField.Equal(list.GetAt(2)))
	require.True(t, s.StringField.Equal(list.GetAt(3)))
	require.True(t, s.ListField.Equal(list.GetAt(4)))
	require.True(t, s.MapField.Equal(list.GetAt(5)))

	innerObj := list.GetAt(100)
	require.NotNil(t, innerObj)
	require.Equal(t, value.LIST, innerObj.Kind())
	innerList := innerObj.(value.List)
	require.True(t, s.InnerField.String.Equal(innerList.GetAt(1)))

}

type RepExample struct {

	BoolField       []value.Bool      `tag:"1" repeated:"true"`
	NumberField     []value.Number    `tag:"2" repeated:"true"`
	StringField     []value.String	  `tag:"3" repeated:"true"`
	ListField       []value.List      `tag:"4" repeated:"true"`
	MapField        []value.Map       `tag:"5" repeated:"true"`
	InnerField      []*Inner          `tag:"100" repeated:"true"`

}


func TestRepStruct(t *testing.T) {

	inner := &Inner {
		String: value.Utf8("inner"),
	}

	s := RepExample{
		BoolField: []value.Bool {value.True, value.False},
		NumberField: []value.Number { value.Long(123), value.Long(456) },
		StringField: []value.String { value.Utf8("test"), value.Raw([]byte("bytes"), false) },
		ListField: []value.List { value.EmptyList(), value.EmptyList() },
		MapField: []value.Map { value.EmptyMap(), value.EmptyMap() },
		InnerField: []*Inner { inner, inner },
	}

	blob, err := value.PackStruct(&s)
	require.Nil(t, err)

	println(hex.EncodeToString(blob))

	var d RepExample
	err = value.UnpackStruct(blob, &d, false)
	require.Nil(t, err)

}


type ArrayExample struct {

	BoolField       []value.Bool      `tag:"1"`
	NumberField     []value.Number    `tag:"2"`
	StringField     []value.String	  `tag:"3"`
	ListField       []value.List      `tag:"4"`
	MapField        []value.Map       `tag:"5"`
	InnerField      []*Inner          `tag:"100"`

}

func TestArrayStruct(t *testing.T) {

	inner := &Inner {
		String: value.Utf8("inner"),
	}

	s := ArrayExample{
		BoolField: []value.Bool {value.True, value.False},
		NumberField: []value.Number { value.Long(123), value.Long(456) },
		StringField: []value.String { value.Utf8("test"), value.Raw([]byte("bytes"), false) },
		ListField: []value.List { value.EmptyList(), value.EmptyList() },
		MapField: []value.Map { value.EmptyMap(), value.EmptyMap() },
		InnerField: []*Inner { inner, inner },
	}

	blob, err := value.PackStruct(&s)
	require.Nil(t, err)

	println(hex.EncodeToString(blob))

	var d ArrayExample
	err = value.UnpackStruct(blob, &d, false)
	require.Nil(t, err)

}
